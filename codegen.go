package qrcgen

import (
    "bytes"
    "github.com/golang/freetype"
    "github.com/nfnt/resize"
    "github.com/skip2/go-qrcode"
    "golang.org/x/image/font"
    "image"
    "image/color"
    "image/draw"
    "image/png"
    "io/ioutil"
    "os"
    "path"
)

var (
    ColorFail = color.RGBA{R: 0xAA, G: 0x00, B: 0x00, A: 0xFF}
    ColorPass = color.RGBA{R: 0x00, G: 0xAA, B: 0x00, A: 0xFF}
)

type Step struct {
    Name string
    Flag bool
}

type Client struct {
    Content  string
    Steps    []Step
    Success  bool
    FontFile string
    FontSize float64
    Spacing  float64
    Assets   string
    DPI      float64
    Size     int
}

func NewClient(content string, steps []Step, assets string, success bool) *Client {
    c := &Client{Content: content, Steps: steps, FontFile: path.Join(assets, "default.ttf"), Assets: assets, Success: success}
    c.FontSize = 50
    c.Spacing = 1.28
    c.Size = 2048
    c.DPI = 72<<2
    return c
}

func (c *Client) Create() (*bytes.Buffer, error) {
    qr, err := qrcode.New(c.Content, qrcode.Medium)
    if err != nil {return nil, err}
    qr.DisableBorder = true
    if !c.Success {qr.ForegroundColor = ColorFail} else { qr.ForegroundColor = ColorPass}

    qr.BackgroundColor = color.Transparent
    data, err := qr.PNG(c.Size)
    if err != nil {return nil, err}

    fb, err := ioutil.ReadFile(c.FontFile)
    if err != nil {return nil, err}
    ft, err := freetype.ParseFont(fb)
    if err != nil {return nil, err}

    w, h := c.Size<<1, c.Size
    canvas := image.NewRGBA(image.Rect(0,0,w,h))
    draw.Draw(canvas, canvas.Bounds(), image.Transparent, image.Point{}, draw.Src)

    img, err := png.Decode(bytes.NewReader(data))
    if err != nil {return nil, err}
    draw.Draw(canvas, canvas.Bounds(), img, image.Point{}, draw.Src)
    {
        r,g,b,a := qr.ForegroundColor.RGBA()
        rc := color.RGBA{R: uint8(r>>8), G: uint8(g>>8), B: uint8(b>>8), A: uint8(float64(a>>8) * 0.1)}
        logo := "ok.png"
        if !c.Success { logo = "no.png" }
        if fp, err := os.Open(path.Join(c.Assets, logo)); err == nil {
            defer fp.Close()
            if img, _, err := image.Decode(fp); err == nil {
                if img, ok := img.(*image.NRGBA); ok {
                    for x := 0; x < img.Bounds().Dx(); x++ {
                        for y := 0; y < img.Bounds().Dy(); y++ {
                            v := img.NRGBAAt(x, y)
                            if v.A > 0 { img.Set(x, y, rc) }
                        }
                    }
                }
                s := img.Bounds().Size()
                o := image.Point{X: c.Size + (c.Size - s.X)/2, Y: (c.Size - s.Y)/2}
                draw.Draw(canvas, canvas.Bounds().Add(o), img, image.Point{}, draw.Src)
            }
        }
    }

    cxt := freetype.NewContext()
    cxt.SetClip(canvas.Bounds())
    cxt.SetHinting(font.HintingFull)
    cxt.SetFontSize(c.FontSize)
    cxt.SetDst(canvas)
    cxt.SetDPI(c.DPI)
    cxt.SetFont(ft)

    pt := freetype.Pt(c.Size + int(c.FontSize*2), int(cxt.PointToFixed(c.FontSize)>>6))
    for _, s := range c.Steps {
        if s.Flag {
            cxt.SetSrc(image.NewUniform(ColorPass))
        } else {
            cxt.SetSrc(image.NewUniform(ColorFail))
        }
        _, err := cxt.DrawString(s.Name, pt)
        if err != nil {break}
        pt.Y += cxt.PointToFixed(c.FontSize * c.Spacing)
    }

    b := &bytes.Buffer{}
    img = resize.Resize(uint(c.Size), uint(c.Size/2), canvas, resize.MitchellNetravali)
    return b, png.Encode(b, img)
}