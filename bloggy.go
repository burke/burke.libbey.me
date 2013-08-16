package main

import (
	"bytes"
	"github.com/russross/blackfriday"
	"hash/fnv"
	"html"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"
)

const (
	BuildDir         = "public"
	PostLayout       = "layouts/post.html"
	IndexSource      = "layouts/index.html"
	StaticDir        = "static"
	PostsDir         = "posts"
	PygmentsCacheDir = ".pygments-cache"
)

func main() {
	createBuildDirectories()
	posts := loadPosts()
	buildPosts(posts)
	buildIndex(posts)
	copyStaticAssets()
}

type Post struct {
	Content, Slug, SourcePath, Title, Url string
	Date                                  time.Time
}

type PostSlice []*Post

func (p PostSlice) Get(i int) *Post    { return p[i] }
func (p PostSlice) Len() int           { return len(p) }
func (p PostSlice) Less(i, j int) bool { return p[i].Date.Unix() < p[j].Date.Unix() }
func (p PostSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p PostSlice) Sort()              { sort.Sort(p); sort.Reverse(p) }
func (p PostSlice) Limit(n int) PostSlice {
	if len(p) <= n {
		return p
	} else {
		return p[0:n]
	}
}

func createBuildDirectories() {
	os.MkdirAll(BuildDir, 0755)
	os.MkdirAll(PygmentsCacheDir, 0755)
}

func loadPosts() PostSlice {
	glob := PostsDir + "/*.md"
	dirfiles, _ := filepath.Glob(glob)
	var posts PostSlice
	for _, file := range dirfiles {
		post := loadPost(file)
		posts = append(posts, post)
	}
	posts.Sort()
	return posts
}

func loadPost(path string) *Post {
	post := new(Post)
	post.SourcePath = path
	markdownBasename := filepath.Base(path)
	datePart := markdownBasename[0:10]
	namePart := markdownBasename[11:]
	post.Slug = strings.Replace(namePart, ".md", "", 1)
	post.Url = "/" + post.Slug

	var err error
	post.Date, err = time.Parse("2006-01-02", datePart)
	if err != nil {
		panic(err)
	}

	return post
}

func buildPosts(posts PostSlice) {
	for _, post := range posts {
		buildPost(post)
	}
}

func buildPost(post *Post) {
	source, err := ioutil.ReadFile(post.SourcePath)
	if err != nil {
		panic(err)
	}

	parts := strings.SplitN(string(source), "\n", 3)
	post.Title = parts[0]
	if parts[1] != "---" {
		panic("Improperly formatted post")
	}
	content := parts[2]

	output := blackfriday.MarkdownCommon([]byte(content))
	post.Content = string(output)

	buffer := new(bytes.Buffer)
	ts, err := template.ParseFiles(PostLayout)
	if err != nil {
		panic(err)
	}
	ts.ExecuteTemplate(buffer, filepath.Base(PostLayout), post)

	highlighted := pygmentize(buffer)

	outDir := BuildDir + "/" + post.Slug
	os.MkdirAll(outDir, 0755)
	outFile := outDir + "/index.html"
	ioutil.WriteFile(outFile, highlighted, 0644)
}

func buildIndex(posts PostSlice) {
	buffer := new(bytes.Buffer)
	ts, err := template.ParseFiles(IndexSource)
	if err != nil {
		panic(err)
	}
	indexData := make(map[string]PostSlice)
	indexData["Recent"] = posts
	ts.ExecuteTemplate(buffer, filepath.Base(IndexSource), indexData)

	outFile := BuildDir + "/index.html"
	ioutil.WriteFile(outFile, []byte(buffer.String()), 0644)
}

func copyStaticAssets() {
	cmd := exec.Command("cp", "-rf", StaticDir, BuildDir)
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

var regex = regexp.MustCompile(`(?ms)<pre><code class="(.*?)">(.*?)</code></pre>`)

func pygmentsReplacer(in []byte) []byte {
	hasher := fnv.New32()
	hasher.Write(in)
	sum := hasher.Sum32()
	strsum := strconv.Itoa(int(sum))

	source, err := ioutil.ReadFile(PygmentsCacheDir + "/" + strsum)
	if err == nil {
		return source
	}

	outs := regex.FindSubmatch(in)
	lang := outs[1]
	content := outs[2]
	content = []byte(html.UnescapeString(string(content)))

	cmd := exec.Command("pygmentize", "-fhtml", "-l", string(lang))
	inp, _ := cmd.StdinPipe()
	inp.Write(content)
	inp.Close()

	output, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	ioutil.WriteFile(PygmentsCacheDir+"/"+strsum, output, 0644)

	return output
}

func pygmentize(buffer *bytes.Buffer) []byte {
	output := regex.ReplaceAllFunc(buffer.Bytes(), pygmentsReplacer)
	return output
}
