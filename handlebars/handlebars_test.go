package handlebars

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"testing"
)

func trim(str string) string {
	trimmed := strings.TrimSpace(regexp.MustCompile(`\s+`).ReplaceAllString(str, " "))
	trimmed = strings.Replace(trimmed, " <", "<", -1)
	trimmed = strings.Replace(trimmed, "> ", ">", -1)
	return trimmed
}

func Test_Render(t *testing.T) {
	t.Parallel()
	engine := New("./views", ".hbs")
	if err := engine.Load(); err != nil {
		t.Fatalf("load: %v\n", err)
	}
	// Partials
	var buf bytes.Buffer
	engine.Render(&buf, "index", map[string]interface{}{
		"Title": "Hello, World!",
	})
	expect := `<h2>Header</h2><h1>Hello, World!</h1><h2>Footer</h2>`
	result := trim(buf.String())
	if expect != result {
		t.Fatalf("Expected:\n%s\nResult:\n%s\n", expect, result)
	}
	// Single
	buf.Reset()
	engine.Render(&buf, "errors/404", map[string]interface{}{
		"Title": "Hello, World!",
	})
	expect = `<h1>Hello, World!</h1>`
	result = trim(buf.String())
	if expect != result {
		t.Fatalf("Expected:\n%s\nResult:\n%s\n", expect, result)
	}
}

func Test_Layout(t *testing.T) {
	t.Parallel()
	engine := New("./views", ".hbs")
	if err := engine.Load(); err != nil {
		t.Fatalf("load: %v\n", err)
	}

	var buf bytes.Buffer
	engine.Render(&buf, "index", map[string]interface{}{
		"Title": "Hello, World!",
	}, "layouts/main")
	expect := `<!DOCTYPE html><html><head><title>Main</title></head><body><h2>Header</h2><h1>Hello, World!</h1><h2>Footer</h2></body></html>`
	result := trim(buf.String())
	if expect != result {
		t.Fatalf("Expected:\n%s\nResult:\n%s\n", expect, result)
	}
}

func Test_Empty_Layout(t *testing.T) {
	t.Parallel()
	engine := New("./views", ".hbs")
	if err := engine.Load(); err != nil {
		t.Fatalf("load: %v\n", err)
	}

	var buf bytes.Buffer
	engine.Render(&buf, "index", map[string]interface{}{
		"Title": "Hello, World!",
	}, "")
	expect := `<h2>Header</h2><h1>Hello, World!</h1><h2>Footer</h2>`
	result := trim(buf.String())
	if expect != result {
		t.Fatalf("Expected:\n%s\nResult:\n%s\n", expect, result)
	}
}

func Test_FileSystem(t *testing.T) {
	t.Parallel()
	engine := NewFileSystem(http.Dir("./views"), ".hbs")
	engine.Debug(true)
	if err := engine.Load(); err != nil {
		t.Fatalf("load: %v\n", err)
	}

	var buf bytes.Buffer
	engine.Render(&buf, "index", map[string]interface{}{
		"Title": "Hello, World!",
	}, "layouts/main")
	expect := `<!DOCTYPE html><html><head><title>Main</title></head><body><h2>Header</h2><h1>Hello, World!</h1><h2>Footer</h2></body></html>`
	result := trim(buf.String())
	if expect != result {
		t.Fatalf("Expected:\n%s\nResult:\n%s\n", expect, result)
	}
}

func Test_Reload(t *testing.T) {
	engine := NewFileSystem(http.Dir("./views"), ".hbs")
	engine.Reload(true) // Optional. Default: false

	if err := engine.Load(); err != nil {
		t.Fatalf("load: %v\n", err)
	}

	if err := ioutil.WriteFile("./views/reload.hbs", []byte("after reload\n"), 0644); err != nil {
		t.Fatalf("write file: %v\n", err)
	}
	defer func() {
		if err := ioutil.WriteFile("./views/reload.hbs", []byte("before reload\n"), 0644); err != nil {
			t.Fatalf("write file: %v\n", err)
		}
	}()

	engine.Load()

	var buf bytes.Buffer
	engine.Render(&buf, "reload", nil)
	expect := "after reload"
	result := trim(buf.String())
	if expect != result {
		t.Fatalf("Expected:\n%s\nResult:\n%s\n", expect, result)
	}
}

func Test_Func(t *testing.T) {
	engine := New("./views", ".hbs")
	if err := engine.Load(); err != nil {
		t.Fatalf("load: %v\n", err)
	}
	engine.AddFunc("add", func(arg1 int, arg2 int) string {
		return fmt.Sprintf("%d", arg1+arg2)
	})
	// Single
	var buf bytes.Buffer
	engine.Render(&buf, "func", map[string]interface{}{
		"Title": "Hello, World!",
	})
	expect := `2`
	result := trim(buf.String())
	if expect != result {
		t.Fatalf("Expected:\n%s\nResult:\n%s\n", expect, result)
	}
}

func Test_Func_Reload(t *testing.T) {
	engine := New("./views", ".hbs")
	engine.Reload(true)
	if err := engine.Load(); err != nil {
		t.Fatalf("load: %v\n", err)
	}
	engine.AddFunc("add2", func(arg1 int, arg2 int) string {
		return fmt.Sprintf("%d", arg1+arg2)
	})
	// Load it again after adding function, simulating reload
	if err := engine.Load(); err != nil {
		t.Fatalf("load: %v\n", err)
	}
	//Test adding more function
	engine.AddFunc("add3", func(arg1 int, arg2 int) string {
		return fmt.Sprintf("%d", arg1+arg2)
	})
	var buf bytes.Buffer
	engine.Render(&buf, "funcrel", nil)
	expect := `2`
	result := trim(buf.String())
	if expect != result {
		t.Fatalf("Expected:\n%s\nResult:\n%s\n", expect, result)
	}
}

func Test_Func_Double_NotAllowed(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("AddFunc accepts 2 functionsa of the same name, this is an error!")
		}
	}()

	engine := New("./views", ".hbs")

	//Try adding 2 functions of the same name
	engine.AddFunc("add4", func(arg string) string {
		return arg
	})
	engine.AddFunc("add4", func(arg string) string {
		return arg
	})
}
