# pdf2thumb4lambda

This is a pretty specialized lambda function, but someone else may find it useful...

This code will respond to an S3 creation notification, pull in the PDF(s) identified, render it using pdfium (Chrome's internal PDF viewer implementation), then run the rendered image through pngquant to reduce file size. It is designed for single-page PDFs, but will create a 2x2 or 3x3 "gallery" view of the first 9 pages if multiple pages are present. It will then deposit the resulting PNG image into a separate S3 bucket (substituting the .pdf with .png).

## Standing on the shoulders of giants

I really just glued some cool projects together. Here are some details about that:

* Read more about the PDFium project here: https://pdfium.googlesource.com/pdfium
  * Instead of compiling PDFium myself, binaries are sourced from here: https://github.com/bblanchon/pdfium-binaries
* Pngquant is a pretty amazing software project found here: https://github.com/kornelski/pngquant
* Cross-compilation of CGO referenced libraries from these project is done via the magical Zig compiler wrappers as shown here: https://andrewkelley.me/post/zig-cc-powerful-drop-in-replacement-gcc-clang.html

## Creating your own Lambda function:

Build the .zip files for upload to AWS using `make` in the project root.

* Runtime: Choose "Custom runtime on Amazon Linux 2"
* Configuration:
  * General: You will likely want to increase the Memory or Timeout from the defaults.
  * Triggers: "S3", select your upload bucket, add a ".pdf" suffix and any prefix you may want.
  * Permissions: Ensure that the Lambda's Role can GetObject from the source bucket and PutObject to the destination bucket.
  * Environment Variables: Add a variable named "DESTINATION_BUCKET" and set it to the S3 bucket used to deposit PNG thumbnails.

## License

The code unique to this repo is released under an MIT license, other included code maintains it's existing license.