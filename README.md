# media-reorg

Command line tool to reorganise a collection of pictures and movies into a set of folders based on year and month.  Also tracks and handles duplicates and other files.

Organises by year/month in the form `2026/09`.  Files are *moved* into that structure inside the source folder so they retain their original timestamps.  For duplicates (by hash) only the earliest is moved into the expected structure.

Remaining duplicates, files with read errors, and files of other filetypes are moved into parallel folder trees for manual checks.

- Copyright K Cartlidge, 2026
- [AGPLv3 license](LICENSE.txt)

## Contents

- [Usage](#usage)
- [Process](#process)
- [Outcome](#outcome)

## Usage

``` bash
cd <repo>
cd cmd
go run . <folder>
```

## Process

- Drop all your media into a single folder (nested how you want, if at all)
- Run *media-reorg* against that folder
- Deal with any duplicates, errors, or other file types

*Note that as the files are rearranged based on date if you already have some categorised in a different folder structure you may not want to run this against them as you will lose that custom structure.*

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
_rm_issues
  duplicates/
    2026/
      06/
        bananas.jpg
        bananas_1.jpg
      12/
        apples.mp4
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
- The top level `_rm_issues` folder contains things to look at manually
  - The `duplicates` are the *later* copies of files whose earliest instance was *kept* in the main dated folders, keeping their original names unless that name is already taken in the destination folder, in which case an incrementing suffix is added (you can safely remove the `duplicates` folder unless you're curious, given that the earliest is in the main collection anyway)
  - The `errors` structure will be populated with errored files, grouped by error then by year/month
  - The `other_filetypes` structure will be populated with just like the main folders but with non-media files
- In reality the issues folder usually lists first; it's last here just for clarity

OS metadata files such as `.DS_Store`, `Thumbs.db`, `desktop.ini`, and AppleDouble `._*` files are skipped entirely and are not organised.

After the files have been moved, empty folders and folders that contain only those junk files are removed.  The source folder itself is left in place.

The `_rm_issues` folder is skipped when scanning (the number of files is reported).  To reprocess any of those files, move them out of that folder tree first.
