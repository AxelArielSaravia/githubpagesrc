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

func build(
    path, tempName string,
    tempFiles []string,
    data any,
) (err error) {
    var home *os.File
    home, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
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

func minifyHTML(source string) (err error) {
    var f, temp *os.File

    temp, err = os.Create(source+".temp.html")
    if err != nil {
        temp, err = os.OpenFile(source+".temp.html", os.O_WRONLY|os.O_APPEND, 0644)
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
    var stylemode bool = false
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
                if slices.Compare(subline, []byte("<style>")) == 0 {
                    stylemode = true
                }
                if slices.Compare(subline, []byte("</style>")) == 0 {
                    stylemode = false
                }

                endbyte := line[len(line)-1]
                if endbyte == '>' {
                    addSpace = false
                }


            } else if !stylemode &&
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

func main() {
    var argsCount = 1
    var argsLen int = len(os.Args)
    if slices.Contains(os.Args[1:], "help") {
        fmt.Print(
            "Usage: <this> [OPTIONS]\n",
            "\n",
            "Options:\n",
            "    prod   production (adds production data)\n",
            "    mini   minify html files\n",
            "\n",
            "    home   build home html file on ../build/\n",
            "    dev    build dev html file on ../build/dev/\n",
            "    music  build music html file on ../build/music/\n",
            "\n",
        )
        return
    }

    var data Data
    if argsLen > argsCount && slices.Contains(os.Args[1:], "prod") {
        argsCount += 1
        data = Data{OriginURL: "https://axelarielsaravia.github.io/"}
    } else {
        data = Data{OriginURL: "/"}
    }

    var miniArg = false
    if argsLen > argsCount && slices.Contains(os.Args[1:], "mini") {
        argsCount += 1
        miniArg = true
    }

    if argsLen < argsCount + 1 || slices.Contains(os.Args[1:], "home") {
        err := build(
            "../build/index.html",
            "home.page.html",
            []string{
                "./templates/base.layout.html",
                "./templates/home.page.html",
            },
            data,
        )
        if err != nil {
            panic(err)
        }
        fmt.Println("Home was successfully created");

        if miniArg {
            err = minifyHTML("../build/index.html")
            if err != nil {
                panic(err)
            }
            fmt.Println("Home was successfully minified");
        }
    }
    if argsLen < argsCount + 1 || slices.Contains(os.Args[1:], "dev") {
        _, err := os.Stat("../build/dev/")
        if err != nil {
            err = os.Mkdir("../build/dev", 0750)
            if err != nil {
                panic(err)
            }
        }
        err = build(
            "../build/dev/index.html",
            "dev.page.html",
            []string{
                "./templates/base.layout.html",
                "./templates/dev.page.html",
            },
            data,
        )
        if err != nil {
            panic(err)
        }
        fmt.Println("dev was successfully created");

        if miniArg {
            err = minifyHTML("../build/dev/index.html")
            if err != nil {
                panic(err)
            }
            fmt.Println("dev was successfully minified");
        }
    }
    if argsLen < argsCount + 1 || slices.Contains(os.Args[1:], "music") {
        _, err := os.Stat("../build/music/")
        if err != nil {
            err = os.Mkdir("../build/music/", 0750)
            if err != nil {
                panic(err)
            }
        }
        err = build(
            "../build/music/index.html",
            "music.page.html",
            []string{
                "./templates/base.layout.html",
                "./templates/music.page.html",
            },
            data,
        )
        if err != nil {
            panic(err)
        }
        fmt.Println("music was successfully created");

        if miniArg {
            err = minifyHTML("../build/music/index.html")
            if err != nil {
                panic(err)
            }
            fmt.Println("music was successfully minified");
        }
    }
    if argsLen < argsCount + 1 || slices.Contains(os.Args[1:], "music") {
        err := build(
            "../build/404.html",
            "404.page.html",
            []string{
                "./templates/base.layout.html",
                "./templates/404.page.html",
            },
            data,
        )
        if err != nil {
            panic(err)
        }
        fmt.Println("404 was successfully created");

        if miniArg {
            err = minifyHTML("../build/404.html")
            if err != nil {
                panic(err)
            }
            fmt.Println("404 was successfully minified");
        }
    }
}
