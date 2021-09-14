package main

import (
    "flag"
    "github.com/larryhou/qrcgen"
    "math/rand"
    "os"
    "strings"
    "time"
)

func main() {
    var content,steps,name,assets string
    result := false
    flag.StringVar(&content, "c", "test", "qrcode content")
    flag.StringVar(&steps, "s", "登录失败", "steps show in right")
    flag.BoolVar(&result, "r", false, "smoke test result flag")
    flag.StringVar(&name, "o", "qrcode.png", "output file name")
    flag.StringVar(&assets, "a", "assets", "assets path")
    flag.Parse()

    rand.Seed(time.Now().UnixNano())

    var items []qrcgen.Step
    for _, s := range strings.Split(steps, ";") {
        items = append(items, qrcgen.Step{Name: s, Flag: rand.Int() % 2 == 1})
    }

    c := qrcgen.NewClient(content, items, assets, result)
    if b, err := c.Create(); err == nil {
        if fp, err := os.OpenFile(name, os.O_CREATE | os.O_WRONLY, 0700); err != nil {panic(err)} else {
            if _, err := fp.Write(b.Bytes()); err != nil {panic(err)}
        }
    } else {panic(err)}
}