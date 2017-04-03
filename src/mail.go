package main

import (
    "github.com/mailgun/mailgun-go"
    "fmt"
    "time"
    "github.com/docker/go-units"
    "github.com/asaskevich/govalidator"
    "errors"
    html_template "html/template"
    "bytes"
)

// https://mailgun.com/app/dashboard
// mail("maarten@lukin.be", "readme.md", "https://push.kiwi/sdfsdf/readme.md", 1024, t)
func mail(recipient string, filename string, url string, size float64, deleteOn time.Time) (err error) {
    if govalidator.IsEmail(recipient) {
        mg := mailgun.NewMailgun(config.MailgunDomain, config.MailgunKey, config.MailgunPublicKey)
        message := mailgun.NewMessage(
            "Push.Kiwi <noreply@push.kiwi>",
            fmt.Sprintf("File received: %s (via Push.Kiwi)", filename),
            fmt.Sprintf("Download link: %s", url),
            recipient,
        )

        tpl_data, err := Asset("static/email.html")
        template, err := html_template.New("email").Parse(string(tpl_data[:]))
        if err != nil {
            return err
        }

        //https://golang.org/src/time/format.go
        const layout = "Monday 2 January, 2006 at 15:04 (MST)"
        data := struct {
            Url         string
            HumanSize   string
            DeletedOn   string
        }{
            url,
            units.HumanSize(size),
            deleteOn.UTC().Format(layout),
        }

        var b bytes.Buffer
        if err := template.Execute(&b, data); err != nil {
            return err
        }

        message.SetHtml(b.String())
        _, _, err = mg.Send(message)
        return err
    }

    return errors.New("Invalid Email")
}
