// this file is used as a workaround, so go <somecommand> ./...
// doesn't recurse into this directory (and mainly the linux dir)
// because it may contain go files, or c files, which go assumes is
// part of your project as well.
// 
// This issue explains this hack (https://github.com/golang/go/issues/30058)
// It's a hack because there's no standard (like .goignore) yet.
// The issue explains it should be empty, but by putting this
// text in here jetbrains will like it as well.
module fake_go_module

go 1.15
