package webassets_test

import (
	"archive/zip"
	"bytes"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi/webassets"
	"github.com/JochiRaider/cartulary/internal/testutil/browserbuild"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
)

func TestReadBrowserRootHTMLValidatesAndInjectsBuildBoundSupport(t *testing.T) {
	t.Run("client support byte ceiling precedes parsing", func(t *testing.T) {
		files := browserbuild.Files(t)
		files["assets/generated/client-extension-support-registry.json"].Data = bytes.Repeat([]byte(" "), (1<<20)+1)
		browserbuild.Seal(t, files)
		if _, err := webassets.Load(files); err == nil || !strings.Contains(err.Error(), "byte limit") {
			t.Fatalf("support byte limit was not applied before parsing: %v", err)
		}
	})

	b := browserbuild.New(t)
	if !strings.Contains(string(b.RootHTML()), `id="cartulary-client-extension-support-registry"`) {
		t.Fatal("missing client support bootstrap")
	}
	for _, mutate := range []func(fstest.MapFS){
		func(f fstest.MapFS) { delete(f, "assets/generated/extension-registry-integrity.json") },
		func(f fstest.MapFS) {
			f["assets/generated/extension-registry-integrity.sha256"].Data = []byte(strings.Repeat("a", 64) + "\n")
		},
		func(f fstest.MapFS) {
			f["assets/generated/client-extension-support-registry.json"].Data = []byte("{}\n")
		},
		func(f fstest.MapFS) { f["assets/generated/client-bootstrap-template.json"].Data = []byte("{}\n") },
	} {
		f := browserbuild.Files(t)
		mutate(f)
		browserbuild.Seal(t, f)
		if _, err := webassets.Load(f); err == nil {
			t.Fatal("corrupt final package admitted")
		}
	}
}
func TestReadBrowserRootHTMLRejectsStaticMappingMismatch(t *testing.T) {
	for _, name := range []string{"assets/extra.js", "index.html", "outside.js", "assets/app.js"} {
		t.Run(name, func(t *testing.T) {
			f := browserbuild.Files(t)
			var data bytes.Buffer
			w := zip.NewWriter(&data)
			for _, entry := range []string{"assets/app.js", name} {
				file, err := w.Create(entry)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := file.Write([]byte("console.log('fixture');")); err != nil {
					t.Fatal(err)
				}
			}
			if err := w.Close(); err != nil {
				t.Fatal(err)
			}
			f["assets/generated/web-assets.zip"].Data = data.Bytes()
			browserbuild.Seal(t, f)
			if _, err := webassets.Load(f); err == nil {
				t.Fatal("invalid static mapping admitted")
			}
		})
	}
}
func TestReadIndexHTMLUsesArchiveWhenPresent(t *testing.T) {
	f := browserbuild.Files(t)
	b, err := webassets.Load(f)
	if err != nil {
		t.Fatal(err)
	}
	// Mutating callers' buffers and returned metadata cannot alter the build.
	original := string(b.RootHTML())
	root := b.RootHTML()
	root[0] = 'x'
	c := b.Contracts()
	c.ClientSupportJSON[0] = 'x'
	f["assets/generated/web-assets.zip"].Data[0] = 'x'
	if string(b.RootHTML()) != original {
		t.Fatal("mutable bootstrap")
	}
	if _, err := fs.ReadFile(b.StaticFS(), "assets/app.js"); err != nil {
		t.Fatal(err)
	}
	if b.Contracts().ClientSupportJSON[0] != '{' {
		t.Fatal("mutable contracts")
	}
}
func TestStaticFSUsesArchiveWhenPresent(t *testing.T) {
	b := browserbuild.New(t)
	body, err := fs.ReadFile(b.StaticFS(), "assets/app.js")
	if err != nil || string(body) != "console.log('fixture');" {
		t.Fatalf("static bytes: %q %v", body, err)
	}
	for _, name := range []string{"index.html", "client-bootstrap-template.json", "client-extension-support-registry.json"} {
		if _, err := fs.ReadFile(b.StaticFS(), name); err == nil {
			t.Fatalf("private metadata served: %s", name)
		}
	}
}
func TestReadIndexHTMLRejectsMissingFinalPackage(t *testing.T) {
	if _, err := webassets.Load(fstest.MapFS{"assets/fallback/index.html": {Data: []byte("fallback")}}); err == nil {
		t.Fatal("missing final package admitted")
	}
}
