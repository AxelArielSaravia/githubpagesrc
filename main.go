package main

import (
    "bufio"
    "fmt"
    "html/template"
    "os"
    "slices"
)

type Data struct{
    OriginURL string
}


func minifyHTML(source string) (err error) {
    var f, temp *os.File

    temp, err = os.Create(source+".temp.html")
    if err != nil {
        temp, err = os.OpenFile(source+".temp.html", os.O_WRONLY|os.O_APPEND, 0o644)
        if err != nil {
            return err
        }
        err = temp.Truncate(0)
        if err != nil {
            return err
        }
    }

    f, err = os.Open(source)
    if err != nil {
        return err
    }

    var r *bufio.Reader = bufio.NewReader(f)
    var othermode bool = false
    var addSpace bool = true
    for {
        line, err := r.ReadBytes('\n')
        if err != nil {
            break;
        }
        for i := 0; i < len(line)-1; i += 1 {
            if line[i] == ' ' || line[i] == '\t' || line[i] == '\n' {
                continue;
            }
            subline := line[i:len(line)-1]
            if line[i] == '<' {
                if slices.Compare(subline, []byte("<style>")) == 0 &&
                slices.Compare(subline, []byte("<script type=\"modal\">")) == 0 {
                    othermode = true
                } else if slices.Compare(subline, []byte("</style>")) == 0 &&
                slices.Compare(subline, []byte("</script>")) == 0 {
                    othermode = false
                } else {
                    endbyte := line[len(line)-1]
                    if endbyte == '>' {
                        addSpace = false
                    }
                }
            } else if !othermode &&
            line[i] != '>' &&
            slices.Compare(line[i:i+1], []byte("/>")) != 0 {

                if addSpace {
                    temp.Write([]byte(" "))
                } else {
                    addSpace = true
                }
            }
            temp.Write(subline)
            break;
        }
    }
    err = f.Close();
    if err != nil {
        return err
    }

    err = temp.Close()
    if err != nil {
        return err
    }

    os.Remove(source)
    os.Rename(source+".temp.html", source)
    return nil
}

func build(
    path, tempName string,
    tempFiles []string,
    data any,
) (err error) {
    var home *os.File
    home, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o644)
    if err != nil {
        return err
    }
    defer home.Close()

    var homeTmpl *template.Template
    homeTmpl, err = template.New(tempName).ParseFiles(tempFiles...)
    if err != nil {
        return err
    }

    err = home.Truncate(0)
    if err != nil {
        return err
    }

    err = homeTmpl.Execute(home, data)
    if err != nil {
        return err
    }

    return nil
}

const (
    ARG_HELP = iota
    ARG_PROD
    ARG_MINI
    ARG_HOME
    ARG_MUSIC
    ARG_DEV
    ARG_404
    ARG_LEN
)

var args = [ARG_LEN]string{
    ARG_HELP:   "help",
    ARG_PROD:   "prod",
    ARG_MINI:   "mini",
    ARG_HOME:   "home",
    ARG_MUSIC:  "music",
    ARG_DEV:    "dev",
    ARG_404:    "404",
}
var helpMessages = [ARG_LEN]string{
    ARG_HELP:   "this text",
    ARG_PROD:   "adds production data",
    ARG_MINI:   "minify html files",
    ARG_HOME:   "build file",
    ARG_MUSIC:  "build file",
    ARG_DEV:    "build file",
    ARG_404:    "build file",
}

const (
    BUILDER_HOME = iota
    BUILDER_MUSIC
    BUILDER_DEV
    BUILDER_404
    BUILDER_LEN
)

var builders = [BUILDER_LEN]struct{
    arg string
    dir string
    dest string
    src string
    templates []string
}{
    BUILDER_HOME: {
        arg: args[ARG_HOME],
        dir: "",
        dest: "../build/index.html",
        src: "home.page.html",
        templates: []string{
            "./templates/base.layout.html",
            "./templates/home.page.html",
        },
    },
    BUILDER_MUSIC: {
        arg: args[ARG_MUSIC],
        dir: "../build/music/",
        dest: "../build/music/index.html",
        src: "music.page.html",
        templates: []string{
            "./templates/base.layout.html",
            "./templates/music.page.html",
        },
    },
    BUILDER_DEV: {
        arg: args[ARG_DEV],
        dir: "../build/dev/",
        dest: "../build/dev/index.html",
        src: "dev.page.html",
        templates: []string{
            "./templates/base.layout.html",
            "./templates/dev.page.html",
        },
    },
    BUILDER_404: {
        arg: args[ARG_404],
        dir: "",
        dest: "../build/404.html",
        src: "404.page.html",
        templates: []string{
            "./templates/base.layout.html",
            "./templates/404.page.html",
        },
    },
}

func main() {
    var Args []string = os.Args[1:]
    if slices.Contains(Args, args[ARG_HELP]) {
        fmt.Print(
            "Usage: "+os.Args[0]+" [OPTIONS]\n",
            "\n",
            "Options:\n",
        )
        for i := range ARG_LEN {
            fmt.Println("\t",args[i],"\t",helpMessages[i])
        }
        fmt.Print(
            "\n",
            "If no build file option added, all files will be build\n",
            "\n",
        )
        return
    }

    var data Data = Data{OriginURL: "/"}

    if argsLen := len(Args); argsLen > 0 {
        var i int = slices.Index(Args[:], args[ARG_PROD])
        if i != -1 {
            data = Data{OriginURL: "https://axelarielsaravia.github.io/"}
            fmt.Println("[[Production]] data")

            if argsLen > 1 {
                Args[0], Args[i] = Args[i], Args[0]
            }
            Args = Args[1:]
        }
    }

    var miniArg = false
    if argsLen := len(Args); argsLen > 0 {
        var i int = slices.Index(Args[:], args[ARG_MINI])
        if i != -1 {
            miniArg = true
            if argsLen > 1 {
                Args[0], Args[i] = Args[i], Args[0]
            }
            Args = Args[1:]
        }
    }

    var buildAll bool = false
    if len(Args) == 0 {
        buildAll = true
    }

    builderIdx := 0
    for builderIdx < BUILDER_LEN {
        builder := builders[builderIdx]
        builderIdx += 1

        if !buildAll {
            if (len(Args) == 0) {
                break
            }
            if j := slices.Index(Args, builder.arg); j == -1 {
                continue
            } else {
                Args[j],Args[len(Args)-1] = Args[len(Args)-1], Args[j]
                Args = Args[:len(Args)-1]
            }
        }
        if builder.dir != "" {
            _, err := os.Stat(builder.dir)
            if err != nil {
                err = os.Mkdir(builder.dir, 0o750)
                if err != nil {
                    panic(err)
                }
            }
        }
        err := build(builder.dest, builder.src, builder.templates, data)
        if err != nil {
            panic(err)
        }
        fmt.Println(builder.arg, "was successfully created");

        if miniArg {
            err = minifyHTML(builder.dest)
            if err != nil {
                panic(err)
            }
            fmt.Println(builder.arg, "was successfully minified");
        }
    }
}
