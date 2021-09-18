# rm-pdf-tools - PDF tools for reMarkable

`rm-pdf-tools` is currently in a very early version, bugs are to be expected. Furthermore,
the intended use case for this tool is to be run 24/7 on a computer,
and your tablet requires an active internet connection.

Right now, this tools probably requires more technical knowledge than hacks, however you do not run _any_ risk
of voiding your warranty or bricking your device.

## Features 
`rm-pdf-tools` adds the following features to your reMarkable tablet (with an active internet connection):
- Add blank pages to annotated PDFs 
- Remove pages from annotated PDFs 

### Demo 
See [here](https://www.reddit.com/r/RemarkableTablet/comments/pqod77/introducing_rmpdftools_insert_pages_and_delete/) for a demo.

## Installation 
### reMarkable
In the root (top-level) directory of your reMarkable cloud, create the following directories:
```
/pdf-tools/
/pdf-tools/work/
/pdf-tools/original/
/pdf-tools/processed/
```

### Server 
To clone the repository and build the binary, run 
```
git clone https://github.com/skius/rm-pdf-tools
cd rm-pdf-tools
go build .
```
Then start the service using `./start.sh` and follow the instructions to authenticate `rm-pdf-tools` with your
reMarkable cloud (courtesy of [rmapi](https://github.com/juruen/rmapi)).

## Usage 

To add/delete pages of a PDF, simply create a folder in `/pdf-tools/work/`
with a name following the [actions format](#Actions-format) corresponding to what you want to change about the PDF.

Then move your PDF into that folder and wait for a few seconds. If everything worked correctly, you should
now find the processed PDF with your changes in the folder `/pdf-tools/processed/`.  
If you accidentally deleted  too much, or still need the original for other reasons,
you can find it in `/pdf-tools/original/`.

See [the demo](resources/demo.mp4) for an example workflow.

### Actions format

The title of the folder you're creating in `work/` should be a comma-separated list of `action`'s.  
An action can be:
- `XaY`: insert `X` pages **a**fter page `Y`
- `XbY`: insert `X` pages **b**efore page `Y`
- `-Y`: delete page `Y`

Note that your title may not contain multiple references to the same page `Y`, e.g., `-3,1a3` is not allowed.  
Also note that the page numbers always refer to the pages of the original document, i.e. `1a1,-2` deletes the original
2nd page, not the freshly inserted page 2.

#### Examples 
- `2a1,-3`: insert 2 pages after page 1, and delete page 3
- `-10,1a1,1b2`: delete page 10, insert 1 page after page 1, and insert 1 page before page 2
- `-1`: delete page 1
