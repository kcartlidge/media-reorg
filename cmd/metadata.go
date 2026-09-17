package main

import (
	"encoding/binary"
	"os"
	"strings"
	"time"
)

// imageCaptureTime returns an embedded capture timestamp for jpeg/jpg/png when present.
// Failures are silent: ok is false and the caller should keep its existing time.
func imageCaptureTime(path, ext string) (time.Time, bool) {
	switch strings.ToLower(ext) {
	case "jpg", "jpeg", "png":
	default:
		return time.Time{}, false
	}

	// one read of the whole file; EXIF is not at a fixed offset in these formats
	data, err := os.ReadFile(path)
	if err != nil {
		return time.Time{}, false
	}

	switch strings.ToLower(ext) {
	case "jpg", "jpeg":
		return jpegCaptureTime(data)
	case "png":
		return pngCaptureTime(data)
	default:
		return time.Time{}, false
	}
}

// jpegCaptureTime finds an APP1 Exif segment and reads a capture date from it
func jpegCaptureTime(data []byte) (time.Time, bool) {
	if len(data) < 4 || data[0] != 0xff || data[1] != 0xd8 {
		return time.Time{}, false
	}

	i := 2
	for i+4 <= len(data) {
		// markers are FF xx; skip padding FFs
		if data[i] != 0xff {
			return time.Time{}, false
		}
		for i < len(data) && data[i] == 0xff {
			i++
		}
		if i >= len(data) {
			return time.Time{}, false
		}
		marker := data[i]
		i++

		// no payload on these standalone markers
		if marker == 0xd8 || marker == 0xd9 || (marker >= 0xd0 && marker <= 0xd7) {
			if marker == 0xd9 {
				return time.Time{}, false
			}
			continue
		}

		// SOS begins compressed image data; APP markers are before this
		if marker == 0xda {
			return time.Time{}, false
		}

		if i+2 > len(data) {
			return time.Time{}, false
		}
		segLen := int(binary.BigEndian.Uint16(data[i:]))
		i += 2
		if segLen < 2 || i+(segLen-2) > len(data) {
			return time.Time{}, false
		}
		payload := data[i : i+(segLen-2)]
		i += segLen - 2

		// APP1 Exif
		if marker == 0xe1 && len(payload) > 6 &&
			payload[0] == 'E' && payload[1] == 'x' && payload[2] == 'i' && payload[3] == 'f' &&
			payload[4] == 0 && payload[5] == 0 {
			if t, ok := exifCaptureTime(payload[6:]); ok {
				return t, true
			}
		}
	}
	return time.Time{}, false
}

// pngCaptureTime finds an eXIf chunk and reads a capture date from it
func pngCaptureTime(data []byte) (time.Time, bool) {
	sig := []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}
	if len(data) < 8 {
		return time.Time{}, false
	}
	for j := 0; j < 8; j++ {
		if data[j] != sig[j] {
			return time.Time{}, false
		}
	}

	i := 8
	for i+12 <= len(data) {
		n := int(binary.BigEndian.Uint32(data[i:]))
		if n < 0 || i+12+n > len(data) {
			return time.Time{}, false
		}
		typ := string(data[i+4 : i+8])
		chunk := data[i+8 : i+8+n]

		if typ == "eXIf" {
			tiff := chunk
			if len(chunk) > 6 &&
				chunk[0] == 'E' && chunk[1] == 'x' && chunk[2] == 'i' && chunk[3] == 'f' &&
				chunk[4] == 0 && chunk[5] == 0 {
				tiff = chunk[6:]
			}
			return exifCaptureTime(tiff)
		}
		if typ == "IEND" {
			return time.Time{}, false
		}
		i += 12 + n
	}
	return time.Time{}, false
}

// exifCaptureTime reads DateTimeOriginal / DateTimeDigitized / DateTime from a TIFF EXIF blob
func exifCaptureTime(tiff []byte) (time.Time, bool) {
	if len(tiff) < 8 {
		return time.Time{}, false
	}

	var bo binary.ByteOrder
	switch {
	case tiff[0] == 'I' && tiff[1] == 'I':
		bo = binary.LittleEndian
	case tiff[0] == 'M' && tiff[1] == 'M':
		bo = binary.BigEndian
	default:
		return time.Time{}, false
	}
	if bo.Uint16(tiff[2:4]) != 42 {
		return time.Time{}, false
	}

	ifd0 := bo.Uint32(tiff[4:8])
	var fallback time.Time
	var haveFallback bool

	// IFD0 DateTime (tag 0x0132) as last-resort fallback
	if s, ok := exifASCIITag(tiff, bo, ifd0, 0x0132); ok {
		if t, ok := parseExifDate(s); ok {
			fallback, haveFallback = t, true
		}
	}

	// ExifIFD pointer (tag 0x8769)
	exifIFD, ok := exifLongTag(tiff, bo, ifd0, 0x8769)
	if ok {
		if s, ok := exifASCIITag(tiff, bo, exifIFD, 0x9003); ok { // DateTimeOriginal
			if t, ok := parseExifDate(s); ok {
				return t, true
			}
		}
		if s, ok := exifASCIITag(tiff, bo, exifIFD, 0x9004); ok { // DateTimeDigitized
			if t, ok := parseExifDate(s); ok {
				return t, true
			}
		}
	}

	return fallback, haveFallback
}

// exifASCIITag returns a trimmed ASCII tag value from an IFD
func exifASCIITag(tiff []byte, bo binary.ByteOrder, ifdOffset uint32, tag uint16) (string, bool) {
	raw, typ, ok := exifTagData(tiff, bo, ifdOffset, tag)
	if !ok || typ != 2 || len(raw) == 0 {
		return "", false
	}
	s := string(raw)
	if i := strings.IndexByte(s, 0); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s), s != ""
}

// exifLongTag returns a LONG tag value from an IFD
func exifLongTag(tiff []byte, bo binary.ByteOrder, ifdOffset uint32, tag uint16) (uint32, bool) {
	raw, typ, ok := exifTagData(tiff, bo, ifdOffset, tag)
	if !ok || typ != 4 || len(raw) < 4 {
		return 0, false
	}
	return bo.Uint32(raw), true
}

// exifTagData finds a tag in an IFD and returns its value bytes and type
func exifTagData(tiff []byte, bo binary.ByteOrder, ifdOffset uint32, want uint16) ([]byte, uint16, bool) {
	if int(ifdOffset)+2 > len(tiff) {
		return nil, 0, false
	}
	n := int(bo.Uint16(tiff[ifdOffset:]))
	entry := int(ifdOffset) + 2
	for i := 0; i < n; i++ {
		if entry+12 > len(tiff) {
			return nil, 0, false
		}
		tag := bo.Uint16(tiff[entry:])
		typ := bo.Uint16(tiff[entry+2:])
		count := bo.Uint32(tiff[entry+4:])
		if tag == want {
			size, ok := exifTypeSize(typ)
			if !ok {
				return nil, 0, false
			}
			total := int(count) * size
			if total < 0 || total > len(tiff) {
				return nil, 0, false
			}
			if total <= 4 {
				return tiff[entry+8 : entry+8+total], typ, true
			}
			off := int(bo.Uint32(tiff[entry+8:]))
			if off < 0 || off+total > len(tiff) {
				return nil, 0, false
			}
			return tiff[off : off+total], typ, true
		}
		entry += 12
	}
	return nil, 0, false
}

// exifTypeSize is the size in bytes of one EXIF/TIFF value of the given type
func exifTypeSize(typ uint16) (int, bool) {
	switch typ {
	case 1, 2, 7: // BYTE, ASCII, UNDEFINED
		return 1, true
	case 3: // SHORT
		return 2, true
	case 4, 9: // LONG, SLONG
		return 4, true
	case 5, 10: // RATIONAL, SRATIONAL
		return 8, true
	default:
		return 0, false
	}
}

// parseExifDate parses EXIF "YYYY:MM:DD HH:MM:SS"
func parseExifDate(s string) (time.Time, bool) {
	t, err := time.ParseInLocation("2006:01:02 15:04:05", s, time.Local)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}
