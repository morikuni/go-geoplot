package main

import (
	"fmt"
	"net/http"

	"github.com/morikuni/go-geoplot"
)

func main() {
	tokyoTower := &geoplot.LatLng{
		Latitude:  35.658584,
		Longitude: 139.7454316,
	}
	googleMapIcon := &geoplot.Icon{
		URL: "https://maps.google.com/mapfiles/ms/icons/red-dot.png",
		Size: &geoplot.Size{
			Width:  32,
			Height: 32,
		},
		Anchor: &geoplot.Point{
			X: 16,
			Y: 32,
		},
		PopupAnchor: &geoplot.Point{
			X: 0,
			Y: -32,
		},
	}

	m := &geoplot.Map{
		Center: tokyoTower,
		Zoom:   7,
		Area: &geoplot.Area{
			From: tokyoTower.Offset(-0.1, -0.1),
			To:   tokyoTower.Offset(0.2, 0.2),
		},
	}
	m.AddMarker(&geoplot.Marker{
		LatLng: tokyoTower,
		Popup:  "Hello\nWorld",
		Icon:   googleMapIcon,
	})
	m.AddPolyline(&geoplot.Polyline{
		LatLngs: []*geoplot.LatLng{
			tokyoTower.Offset(-0.1, -0.1),
			tokyoTower.Offset(-0.1, 0.1),
			tokyoTower.Offset(0.1, 0.1),
			tokyoTower.Offset(0.1, -0.1),
			tokyoTower.Offset(-0.1, -0.1),
		},
		Popup: "World",
	})
	err := http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := geoplot.ServeMap(w, r, m)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	fmt.Println(err)
}
