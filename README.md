# media-reorg

Command line tool to reorganise a collection of pictures and movies into a set of folders based on year and month.  Also tracks and handles duplicates and other files.  Optionally connects to (ideally local) AI to rename images according to their content (image recognition).

Organises by year/month in the form `2026/09`.  Files are *moved* into that structure inside the source folder so they retain their original timestamps.  For duplicates (by hash) only the earliest is moved into the expected structure.

Remaining duplicates, files with read errors, and files of other filetypes are moved into parallel folder trees for manual checks.

- Copyright K Cartlidge, 2026
- [AGPLv3 license](LICENSE.txt)

## Contents

- [Usage](#usage)
- [Process](#process)
- [Outcome](#outcome)

## Usage

There are two modes which you run separately (or skip).  Both operate in-place rather than making copies, as full copies often won't fit with large collections on space-constrained drives.  Whilst nothing is ever deleted, you may still want to pause any live backups until you see whether the result suits your needs as moving and/or renaming is irreversible unless you use version control.

- `rearrange` reorganises a target folder and all subfolders into a fixed folder layout as described shortly
- `rename` can connect to an ai api to feed it your images one at a time and rename it to describe the contents
  - using a *local* ai is recommended for privacy reasons
  - otherwise OpenRouter is recommened for the free models

There are reasons for the two modes being separate:

- You may want to keep your existing naming convention and just reorganise the folder structure
- You may want to keep your existing folder structure and just rename your files
- There is a performance overhead using ai which means you may wish to run that as an overnight thing on larger collections

*On a 16GB M4 MacBook Pro renaming is surprisingly fast. Using Gemma 4 E4B running in LM Studio images are read, analysed, and renamed in just under a second each without going to the cloud.*

### `rearrange` mode

To reorganise your media folder, just point it at the top folder of your collection:

``` bash
cd <repo>/cmd
go run . -action rearrange [--add-date-prefix] -folder <folder>
```

This will rearrange the entire folder tree as detailed below.

The optional `--add-date-prefix` will also ensure the media date is at the start of the filename (eg `img00032.png` might become `2026-10-28 img00032.png`).  If there's already a date prefix it will be left alone unless it's wrong.

### `rename` mode

The ai (again, prefer a local one like Gemma 4 E4B) will rename any files that appear to currently have generic names with more relevant filenames. For example it might change `IMG00034.PNG` to `chicken-tomato-pasta-salad.png` if that image contains a picture of ready-meal salad.

You don't need a powerful model.  If you use a model like Gemma 4 E4B then in less than a second on an M4 MacBook Pro 16GB with LM Studio it will assess the image, *including reading any text* (eg labels), and rename the file.

Whilst each run checks all the files, even those that may have been renamed previously, it first looks at the existing filename to see if it appears to already be descriptive as opposed to general.  The image recognition only kicks in if needed.  It's therefore quicker the next time around as many image recognition attempts will be skipped (as an example a run that took 2m58s first time was completed in 25s on a re-run).

To do renames you specify the needed config on the command line:

``` bash
cd <repo>/cmd
go run . -action rename [--add-date-prefix] -folder <folder> -api <url> -model <model> -api-key <api-key>
```

For example:

``` bash
cd <repo>/cmd
go run . -action rename -folder ~/my-images -api http://127.0.0.1:1234 -model google/gemma-4-e4b
```

The `url` must be for an Open AI compatible api and the `model` should be supported (if not the list of known models for the api will be shown). The `api-key` is optional because it's strongly recommended you run a local model (as the model will see all your images) and that won't need one.

Install LM Studio (or Ollama etc) then use it to download the *Gemma 4 E4B* model and run it locally.  It *doesn't* need a graphics card to run, though will obviously be faster with one.  Other models may work fine, but Gemma 4 E4B is small, reliable, fast (vs others on the same hardware), accepts images, and also runs okay on a CPU-only machine.  LM Studio will tell you the API endpoint and exact model name.

The optional `--add-date-prefix` behaves the same as for `rearrange`.

## Process

- Drop all your media into a single folder (nested how you want, if at all)
- Run *media-reorg* against that folder
  - Run *rearrange* if you want to tidy the structure
  - Run *rename* if you want filenames to match contents
- Deal with any duplicates, errors, or other file types

*Note that as the files are rearranged based on date (see below) if you already have some categorised in a different folder structure you may not want to run this against them as you will lose that custom structure.*

## Outcome

Your folder will end up looking like this:

``` text
2013/
  10/
    bananas.jpg
    oranges.png
2026/
  06/
    bananas.jpg
  09/
    apples.mp4
    oranges.png
_mr_issues
  ai-failures/
    2026/
      09/
        IMG_0042.jpg
  duplicates/
    a1b2c3d4.../
      album/
        bananas.jpg
      backup/
        bananas.jpg
  errors/
    failed_to_move_file/
      2026/
        06/
          grapes.jpeg
  other_filetypes/
    2025/
      02/
        thing.txt
```

- The top level dated folders contain the media files moved successfully
- The top level `_mr_issues` folder contains things to look at manually
  - The `duplicates` tree groups by content hash: later copies are *moved* there under their original paths relative to the source folder, and a *copy* of the earliest is placed the same way for reference (the earliest itself remains in the main dated folders).  Name collisions in a destination folder get an incrementing suffix.  You can safely remove the `duplicates` folder unless you're curious, given that the earliest is in the main collection anyway
  - The `errors` structure will be populated with errored files, grouped by error then by year/month
  - The `ai-failures` structure holds images the optional AI pass could not process, grouped by year/month
  - The `other_filetypes` structure will be populated with just like the main folders but with non-media files
- In reality the issues folder usually lists first; it's last here just for clarity

OS metadata files such as `.DS_Store`, `Thumbs.db`, `desktop.ini`, and AppleDouble `._*` files are skipped entirely and are not organised.

After the files have been moved, empty folders and folders that contain only those junk files are removed.  The source folder itself is left in place.

The `_mr_issues` folder is skipped when scanning (the number of files is reported).  To reprocess any of those files, move them out of that folder tree first.
