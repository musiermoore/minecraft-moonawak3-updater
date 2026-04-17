# MoonAwak3 Minecraft Updater

Console updater for the MoonAwak3 Minecraft mod pack. The app downloads the public Yandex Disk archive, unpacks it, lets you choose which archive folders/files should be copied into the local `mods` folder, and then replaces the old `mods` folder.

## Requirements

- Go 1.25 or newer
- Internet access for downloading the archive

## Run

```sh
go run .
```

## Build

Build for the current platform:

```sh
go build .
```

Build a Windows executable:

```sh
GOOS=windows GOARCH=amd64 go build -o updater.exe .
```

## Selection Controls

The selector shows only the supported top-level archive folders:

- `mods`
- `shaderpacks`
- `Новый мод` / `Новые моды`

Default selected folders:

- `mods`
- `Новый мод` / `Новые моды`

Controls:

- `↑` / `↓`: move the cursor
- `1-9`: select or unselect a visible item
- `Space`: select or unselect the highlighted item
- `Shift+1-9`: expand or collapse a top-level folder
- `→`: expand the highlighted top-level folder
- `←`: collapse the expanded top-level folder
- `Enter` or `0`: continue with the current selection
- `Esc`: when a folder is open, return to the root view
- `Esc` twice on the root view: close the app

Only top-level folders can be expanded. When a top-level folder is expanded, the selector shows its direct files and folders.

## Generated Files

The app may create these runtime files and folders:

- `mods.zip`
- `temp_mods/`
- `temp_selected_mods/`
- `mods/`

They are ignored by Git.
