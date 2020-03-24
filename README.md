# goclone
A tool to copy your go project to another folder without the hassle of updating the "import" manually.

### Installation

```
go get github.com/gcsomoza/goclone/...
```

### Parameters

```
-s, --source        Source directory.
-d, --destination   Destination directory.
-o, --overwrite     Overwrite destination directory.
--go-only           Clone .go files only.
```

### Usage

Basic cloning.
```
goclone -s github.com/xxx/your_project -d github.com/xxx/your_copied_project
```

Clone to an existing directory.
```
goclone -s github.com/xxx/your_project -d github.com/xxx/your_copied_project -o
```

Clone only .go files.
```
goclone -s github.com/xxx/your_project -d github.com/xxx/your_copied_project -o --go-only
```
