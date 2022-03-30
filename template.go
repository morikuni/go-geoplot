package geoplot

// language=GoTemplate
const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>go-geoplot</title>
    <link rel="stylesheet" href="https://unpkg.com/leaflet@1.7.1/dist/leaflet.css"
      integrity="sha512-xodZBNTC5n17Xt2atTPuE1HxjVMSvLVW9ocqUKLsCC5CXdbqCmblAshOMAS6/keqq/sMZMZ19scR4PsZChSR7A=="
      crossorigin=""/>
    <script src="https://unpkg.com/leaflet@1.7.1/dist/leaflet.js"
      integrity="sha512-XQoYMqMTK8LvdxXYG3nZ448hOEQiglfqkJs1NOQV44cWnUrBc8PkAOcXy20w0vlaXaVUearIOBhiXZ5V3ynxwA=="
      crossorigin=""></script>
    <style type="text/css">
        html, body, #map {
            height: 100%;
            margin: 0;
        }
    </style>
</head>
<body>
    <div id="map"></div>
    <script>
      const map = L.map('map', {
        center: [0, 0],
        zoom: 1,
      });
      // tiles are copied from: https://leaflet-extras.github.io/leaflet-providers/preview/
	  const Esri_WorldStreetMap = L.tileLayer('https://server.arcgisonline.com/ArcGIS/rest/services/World_Street_Map/MapServer/tile/{z}/{y}/{x}', {
	    attribution: 'Tiles &copy; Esri &mdash; Source: Esri, DeLorme, NAVTEQ, USGS, Intermap, iPC, NRCAN, Esri Japan, METI, Esri China (Hong Kong), Esri (Thailand), TomTom, 2012'
      }).addTo(map);
      const OpenStreetMap_Mapnik = L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        maxZoom: 19,
        attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
      });
      const layers = {
     	"Esri": Esri_WorldStreetMap, 
     	"OpenStreetMap": OpenStreetMap_Mapnik, 
      }
      L.control.layers(layers).addTo(map);
      map.attributionControl.addAttribution('&copy; <a href="https://github.com/morikuni/go-geoplot">go-geoplot</a> by <a href="https://github.com/morikuni">morikuni</a>');
      {{ range .lines }}
      {{- .}}
      {{ end }}
    </script>
</body>
</html>
`
