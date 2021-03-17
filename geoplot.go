package geoplot

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"html/template"
	"image/color"
	"io"
	"net/http"
	"strings"
)

type Map struct {
	Center *LatLng
	Zoom   int
	Area   *Area

	markers   []*Marker
	polylines []*Polyline
	circles   []*Circle
}

func (m *Map) AddMarker(mk *Marker) {
	m.markers = append(m.markers, mk)
}

func (m *Map) AddPolyline(pl *Polyline) {
	m.polylines = append(m.polylines, pl)
}

func (m *Map) AddCircle(c *Circle) {
	m.circles = append(m.circles, c)
}

func (m *Map) toJS() (template.JS, error) {
	var lines []string
	if m.Zoom > 0 {
		lines = append(lines, fmt.Sprintf("map.setZoom(%d);",
			m.Zoom,
		))
	}
	if m.Center != nil {
		lines = append(lines, fmt.Sprintf("map.setView([%f, %f]);",
			m.Center.Latitude,
			m.Center.Longitude,
		))
	}
	if m.Area != nil {
		lines = append(lines, fmt.Sprintf("map.fitBounds([[%f, %f],[%f, %f]]);",
			m.Area.From.Latitude,
			m.Area.From.Longitude,
			m.Area.To.Latitude,
			m.Area.To.Longitude,
		))
	}

	return template.JS(strings.Join(lines, " ")), nil
}

type LatLng struct {
	Latitude  float64
	Longitude float64
}

func (l *LatLng) Offset(lat, lon float64) *LatLng {
	return &LatLng{
		l.Latitude + lat,
		l.Longitude + lon,
	}
}

type Area struct {
	From *LatLng
	To   *LatLng
}

type Marker struct {
	LatLng *LatLng
	Popup  string
	Icon   *Icon
}

func (m *Marker) toJS() (template.JS, error) {
	opt := &bytes.Buffer{}
	opt.WriteString("{")
	if m.Icon != nil {
		fmt.Fprintf(opt, `"icon": %s,`, m.Icon.id)
	}
	opt.WriteString("}")

	return template.JS(fmt.Sprintf("L.marker([%f, %f], %s).addTo(map).bindPopup(%q);",
		m.LatLng.Latitude,
		m.LatLng.Longitude,
		opt.String(),
		strings.Replace(m.Popup, "\n", "<br/>", -1),
	)), nil
}

type Polyline struct {
	LatLngs []*LatLng
	Popup   string
	Color   *color.RGBA
}

func (pl *Polyline) toJS() (template.JS, error) {
	var latlngs []string
	for _, l := range pl.LatLngs {
		latlngs = append(latlngs, fmt.Sprintf("[%f, %f]", l.Latitude, l.Longitude))
	}

	type option struct {
		Color string `json:"color,omitempty"`
	}

	opt := option{}
	if pl.Color != nil {
		opt.Color = fmt.Sprintf("#%02x%02x%02x", pl.Color.R, pl.Color.G, pl.Color.B)
	}

	bs, err := json.Marshal(opt)
	if err != nil {
		return "", err
	}

	return template.JS(fmt.Sprintf("L.polyline(%s, %s).addTo(map).bindPopup(%q);",
		"["+strings.Join(latlngs, ",")+"]",
		string(bs),
		strings.Replace(pl.Popup, "\n", "<br/>", -1),
	)), nil
}

type Circle struct {
	LatLng      *LatLng
	RadiusMeter int
	Popup       string
}

func (c *Circle) toJS() (template.JS, error) {
	return template.JS(fmt.Sprintf("L.circle([%f, %f], {radius: %d}).addTo(map).bindPopup(%q);",
		c.LatLng.Latitude,
		c.LatLng.Longitude,
		c.RadiusMeter,
		strings.Replace(c.Popup, "\n", "<br/>", -1),
	)), nil
}

type Point struct {
	X int
	Y int
}

type Size struct {
	Width  int
	Height int
}

type Icon struct {
	URL         string
	HTML        string
	Size        *Size
	Anchor      *Point
	PopupAnchor *Point

	id string
}

func htmlIcon(r, g, b int) string {
	const format = `
<svg width="100%%" height="100%%" viewBox="0 0 32 48" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" xml:space="preserve" xmlns:serif="http://www.serif.com/" style="fill-rule:evenodd;clip-rule:evenodd;stroke-linejoin:round;stroke-miterlimit:2;">
    <path d="M1.701,23.023C0.613,20.877 0,18.454 0,15.89C0,7.12 7.169,0 15.998,0C24.828,0 31.997,7.12 31.997,15.89C31.997,18.454 31.389,20.854 30.3,23L15.998,48.007L1.701,23.023Z" style="fill:rgb(%d,%d,%d);"/>
    <path d="M1.701,23.023C0.613,20.877 0,18.454 0,15.89C0,7.12 7.169,0 15.998,0C24.828,0 31.997,7.12 31.997,15.89C31.997,18.454 31.389,20.854 30.3,23L15.998,48.007L1.701,23.023ZM2.582,22.549C1.57,20.544 1,18.283 1,15.89C1,7.67 7.722,1 15.998,1C24.274,1 30.997,7.67 30.997,15.89C30.997,18.277 30.434,20.513 29.425,22.514C29.419,22.526 15.998,45.993 15.998,45.993L2.582,22.549Z"/>
    <g transform="matrix(1.02055,0,0,1.02055,-1.48306,1.74407)">
        <circle cx="17.129" cy="13.971" r="5.862" style="fill:white;"/>
    </g>
</svg>
`
	return fmt.Sprintf(format, r, g, b)
}

func ColorIcon(r, g, b int) *Icon {
	return &Icon{
		HTML: htmlIcon(r, g, b),
		Size: &Size{
			Width:  20,
			Height: 30,
		},
		Anchor: &Point{
			X: 10,
			Y: 30,
		},
		PopupAnchor: &Point{
			X: 0,
			Y: -30,
		},
	}
}

func (i *Icon) toJS() (template.JS, error) {
	type icon struct {
		HTML        string `json:"html,omitempty"`
		IconURL     string `json:"iconUrl"`
		IconSize    [2]int `json:"iconSize,omitempty"`
		IconAnchor  [2]int `json:"iconAnchor,omitempty"`
		PopupAnchor [2]int `json:"popupAnchor,omitempty"`
		ClassName   string `json:"className"`
	}

	method := "icon"
	ic := icon{
		IconURL: i.URL,
	}
	if ic.IconURL == "" {
		method = "divIcon"
		ic.HTML = i.HTML
		ic.ClassName = ""
	}
	if i.Size != nil {
		ic.IconSize = [2]int{i.Size.Width, i.Size.Height}
	}
	if i.Anchor != nil {
		ic.IconAnchor = [2]int{i.Anchor.X, i.Anchor.Y}
	}
	if i.PopupAnchor != nil {
		ic.PopupAnchor = [2]int{i.PopupAnchor.X, i.PopupAnchor.Y}
	}

	bs, err := json.Marshal(ic)
	if err != nil {
		return "", err
	}

	return template.JS(fmt.Sprintf("const %s = L.%s(%s);", i.id, method, string(bs))), nil
}

func generateID() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var bs [16]byte
	_, err := io.ReadFull(rand.Reader, bs[:])
	if err != nil {
		panic(err)
	}

	for i := range bs {
		bs[i] = chars[int(bs[i])%len(chars)]
	}

	return string(bs[:])
}

func ServeMap(w http.ResponseWriter, _ *http.Request, m *Map) error {
	tmpl, err := template.New("").Parse(htmlTemplate)
	if err != nil {
		return err
	}

	icons := make(map[string]*Icon)
	for _, mk := range m.markers {
		if mk.Icon == nil {
			continue
		}
		i := mk.Icon
		if i.id == "" {
			i.id = generateID()
		}
		_, ok := icons[i.id]
		if ok {
			continue
		}
		icons[i.id] = i
	}

	var lines []template.JS

	l, err := m.toJS()
	if err != nil {
		return err
	}
	lines = append(lines, l)

	for _, i := range icons {
		l, err := i.toJS()
		if err != nil {
			return err
		}
		lines = append(lines, l)
	}

	for _, c := range m.circles {
		l, err := c.toJS()
		if err != nil {
			return err
		}
		lines = append(lines, l)
	}

	for _, mk := range m.markers {
		l, err := mk.toJS()
		if err != nil {
			return err
		}
		lines = append(lines, l)
	}

	for _, pl := range m.polylines {
		l, err := pl.toJS()
		if err != nil {
			return err
		}
		lines = append(lines, l)
	}

	err = tmpl.Execute(w, map[string]interface{}{
		"lines": lines,
	})
	if err != nil {
		return err
	}

	return nil
}
