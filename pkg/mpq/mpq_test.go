package mpq_test

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/nielsAD/noot/pkg/mpq"
)

func TestMPQ(t *testing.T) {
	archive, err := mpq.OpenArchive("./test.mpq")
	if err != nil {
		t.Fatal("test.mpq", err)
	}

	hello, err := archive.Open("hello.txt")
	if err != nil {
		t.Fatal("hello.txt", err)
	}
	if content, err := ioutil.ReadAll(hello); err != nil {
		t.Fatal("hello.txt", err)
	} else if hello.Size() != int64(len(content)) {
		t.Fatalf("hello.txt: size '%v' != '%v'\n", hello.Size(), len(content))
	} else if strings.TrimSpace(string(content)) != "Hello" {
		t.Fatalf("hello.txt: '%v' != '%v'\n", string(content), "Hello")
	}
	if err := hello.Close(); err != nil {
		t.Fatal("hello.txt", err)
	}

	// Test subfolders
	world, err := archive.Open("sub\\WORLD.txt")
	if err != nil {
		t.Fatal("WORLD.txt", err)
	}
	if content, err := ioutil.ReadAll(world); err != nil {
		t.Fatal("WORLD.txt", err)
	} else if world.Size() != int64(len(content)) {
		t.Fatalf("WORLD.txt: size '%v' != '%v'\n", world.Size(), len(content))
	} else if strings.TrimSpace(string(content)) != "world" {
		t.Fatalf("WORLD.txt: '%v' != '%v'\n", string(content), "world")
	}
	if err := world.Close(); err != nil {
		t.Fatal("WORLD.txt", err)
	}

	// Test case insensivity
	world2, err := archive.Open("SUB\\world.txt")
	if err != nil {
		t.Fatal("world.txt", err)
	}
	if content, err := ioutil.ReadAll(world2); err != nil {
		t.Fatal("world.txt", err)
	} else if world2.Size() != int64(len(content)) {
		t.Fatalf("world.txt: size '%v' != '%v'\n", world2.Size(), len(content))
	} else if strings.TrimSpace(string(content)) != "world" {
		t.Fatalf("world.txt: '%v' != '%v'\n", string(content), "world")
	}
	if err := world2.Close(); err != nil {
		t.Fatal("world.txt", err)
	}

	// Test non existant
	if _, err := archive.Open("foobar.txt"); err != os.ErrNotExist {
		t.Fatal("foobar.txt", err)
	}

	if err := archive.Close(); err != nil {
		t.Fatal("test.mpq", err)
	}
}
