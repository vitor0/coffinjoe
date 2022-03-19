package main

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"text/template"
	"time"

	"github.com/heltonmarx/coffinjoe/schema"
)

// Response represents the request response.
type Response struct {
	XMLName  xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	SoapBody *SOAPBodyResponse
}

// SOAPBodyResponse represents the body of SOAP response
type SOAPBodyResponse struct {
	XMLName xml.Name `xml:"Body"`
	Resp    *ResponseBody
}

// ResponseBody represents the complete body.
type ResponseBody struct {
	XMLName xml.Name `xml:"getObitosResponse"`
	Body    []struct {
		HjidAttr int64 `xml:"Hjid,attr,omitempty"`
		*schema.CObito
	} `xml:"listaObitos"`
}

// Option allows to configure various aspects of Client
type Option func(*Client)

// WithCredentials option which set the username and password
// used as credentials.
func WithCredentials(username, password string) Option {
	return func(c *Client) {
		c.username, c.password = username, password
	}
}

// Client holds the http client used to request
type Client struct {
	conn     *http.Client
	host     string
	username string
	password string
}

// NewClient returns a new http client setup the
// idle connection timeout and timeout
func NewClient(host string, opts ...Option) *Client {
	client := &Client{
		host: host,
		conn: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:    50,
				IdleConnTimeout: 1 * time.Hour,
			},
			Timeout: 5 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(client)
	}
	return client
}

// GetDeatchCertificateByDate get a list of death certificate of given date period.
func (c *Client) GetDeatchCertificateByDate(ctx context.Context, date string) ([]*schema.CObito, error) {
	host, err := url.Parse(c.host)
	if err != nil {
		return nil, err
	}
	buf, err := getRequest(c.username, c.password, date)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, host.String(), bytes.NewBuffer(buf))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/soap+xml;charset=UTF-8")
	resp, err := c.conn.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch certificate, status(%#v)", resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	r := &Response{}
	if err := xml.Unmarshal(removeNonUTF8Bytes(body), &r); err != nil {
		return nil, err
	}
	n := len(r.SoapBody.Resp.Body)
	cobitos := make([]*schema.CObito, n)
	for _, m := range r.SoapBody.Resp.Body {
		cobitos = append(cobitos, m.CObito)
	}
	return cobitos, nil
}

const envelope = `
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" encoding="UTF-16" xmlns:selo="http://www.tjsc.jus.br/selo">
  <soapenv:Header/>
  <soapenv:Body>
    <selo:getObitos>
      <user>{{.Username}}</user>
      <pass>{{.Password}}</pass>
      <data>{{.Date}}</data>
    </selo:getObitos>
  </soapenv:Body>
</soapenv:Envelope>
`

func getRequest(username, password, date string) ([]byte, error) {
	request := struct {
		Username string
		Password string
		Date     string
	}{
		Username: username,
		Password: password,
		Date:     date,
	}
	t, err := template.New("InputRequest").Parse(envelope)
	if err != nil {
		return nil, err
	}
	doc := &bytes.Buffer{}
	if err := t.Execute(doc, request); err != nil {
		return nil, err
	}
	buffer := &bytes.Buffer{}
	encoder := xml.NewEncoder(buffer)
	if err := encoder.Encode(doc.String()); err != nil {
		return nil, err
	}
	return doc.Bytes(), nil
}

func exportJSON(certificates []*schema.CObito, output string) error {
	data, err := json.MarshalIndent(certificates, "", " ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(output+".json", data, 0644)
}

func exportXML(certificates []*schema.CObito, output string) error {
	data, err := xml.MarshalIndent(certificates, "", " ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(output+".xml", data, 0644)
}
