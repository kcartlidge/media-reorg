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

To reorganise your media folder, the simplest option is to do the following:

``` bash
cd <repo>
cd cmd
go run . <folder>
```

### Optional AI renaming

If you want to use an *optional* AI model API to examine image contents and rename the files you can specify the needed config on the command line:

``` bash
cd <repo>
cd cmd
go run . <folder> [<url> <model> [<api-key>]]
```

For example: `~/my-images http://127.0.0.1:1234 google/gemma-4-e4b`

The square brackets show that if you specify the `url` (an Open AI compatible API) you also need to specify the `model` and an optional `api-key`.  The API key is optional because *it's strongly recommended you run a local model as the model will see all your images*.

*Running AI against each image will notably lengthen the time taken.  Normal file reorganisation is completed before the (optional) AI work starts.*

Install LM Studio (or Ollama etc) then use it to download the *Gemma 4 E4B* model and run it locally.  It *doesn't* need a graphics card to run, though will obviously be faster with one.  Other models may work fine, but Gemma 4 E4B is small, reliable, fast (vs others on the same hardware), accepts images, and also runs okay on a CPU-only machine.

LM Studio will tell you the API endpoint and exact model name.  You only need the API key if you are connecting to a cloud model.  If you are still unsure about exact model names run *media-reorg* with random text and it will list the names of those that the API makes available.

With this in place when you run *media-reorg* it will move all the files around as expected but then also feed each *image* file in turn to the API and ask it for a new filename based on the contents (for example `34EC32.png` might become `dog-in-field-sunny-day.png`).  If the current filename doesn't 'seem' random to the model it will leave the name alone.

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
- The top level `_rm_issues` folder contains things to look at manually
  - The `duplicates` tree groups by content hash: later copies are *moved* there under their original paths relative to the source folder, and a *copy* of the earliest is placed the same way for reference (the earliest itself remains in the main dated folders).  Name collisions in a destination folder get an incrementing suffix.  You can safely remove the `duplicates` folder unless you're curious, given that the earliest is in the main collection anyway
  - The `errors` structure will be populated with errored files, grouped by error then by year/month
  - The `other_filetypes` structure will be populated with just like the main folders but with non-media files
- In reality the issues folder usually lists first; it's last here just for clarity

OS metadata files such as `.DS_Store`, `Thumbs.db`, `desktop.ini`, and AppleDouble `._*` files are skipped entirely and are not organised.

After the files have been moved, empty folders and folders that contain only those junk files are removed.  The source folder itself is left in place.

The `_rm_issues` folder is skipped when scanning (the number of files is reported).  To reprocess any of those files, move them out of that folder tree first.
