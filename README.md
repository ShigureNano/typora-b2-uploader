# Typora B2 Uploader

This is a tool for uploading images from Typora to Backblaze B2. By configuring B2 credentials and a custom domain, you can upload images to a specified B2 bucket and generate corresponding URLs.

## Note

You need to configure B2 credentials and a custom domain before using this tool. 

## To build

You need [Go](https://golang.org/dl/) installed, then run:

```bash
go get github.com/kurin/blazer/b2
go build
```
