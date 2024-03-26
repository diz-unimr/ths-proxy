# ths-proxy
![go](https://github.com/diz-unimr/ths-proxy/actions/workflows/build.yml/badge.svg) ![docker](https://github.com/diz-unimr/ths-proxy/actions/workflows/release.yml/badge.svg) [![codecov](https://codecov.io/gh/diz-unimr/ths-proxy/branch/main/graph/badge.svg?token=D66XMZ5ALR)](https://codecov.io/gh/diz-unimr/ths-proxy)
> A small proxy service to intercept webservice requests to TTP tools

This proxy is meant to rewrite certain HTTP requests to TTP tools in order to support custom workflows.

## Routes

Currently, this proxy only supports delegating requests to the gICS SOAP web service [/gics/gicsService](https://www.ths-greifswald.de/wp-content/uploads/tools/gics/doc/2023-1-0/org/emau/icmvc/ganimed/ttp/cm2/GICSService.html) to the 
corresponding endpoint with notifications [/gics/gicsServiceWithNotification](https://www.ths-greifswald.de/wp-content/uploads/tools/gics/doc/2023-1-0/org/emau/icmvc/ganimed/ttp/cm2/GICSServiceWithNotification.html).

The SOAP request body is rewritten to add the `notificationClientID` in this case. Other requests are just forwarded unaltered.