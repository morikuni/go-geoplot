package geoplot

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"html/template"
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
}

func (m *Map) AddMarker(mk *Marker) {
	m.markers = append(m.markers, mk)
}

func (m *Map) AddPolyline(pl *Polyline) {
	m.polylines = append(m.polylines, pl)
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
}

func (pl *Polyline) toJS() (template.JS, error) {
	var latlngs []string
	for _, l := range pl.LatLngs {
		latlngs = append(latlngs, fmt.Sprintf("[%f, %f]", l.Latitude, l.Longitude))
	}

	return template.JS(fmt.Sprintf("L.polyline(%s).addTo(map).bindPopup(%q);",
		"["+strings.Join(latlngs, ",")+"]",
		strings.Replace(pl.Popup, "\n", "<br/>", -1),
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
	Size        *Size
	Anchor      *Point
	PopupAnchor *Point

	id string
}

func (i *Icon) toJS() (template.JS, error) {
	type icon struct {
		IconURL     string `json:"iconUrl"`
		IconSize    [2]int `json:"iconSize,omitempty"`
		IconAnchor  [2]int `json:"iconAnchor,omitempty"`
		PopupAnchor [2]int `json:"popupAnchor,omitempty"`
	}

	ic := icon{
		IconURL: i.URL,
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

	return template.JS(fmt.Sprintf("const %s = L.icon(%s);", i.id, string(bs))), nil
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
