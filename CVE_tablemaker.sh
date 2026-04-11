#!/bin/bash
## Este script requiere de un paso previo, copiar una lista de CVEs de un servicio concreto de https://www.cvedetails.com/ y pegarla en un archivo
## Una vez dicho archivo existe, este script procesa su contenido y genera tablas para pegar en obsidian con CVE | CVSS | Link a NIST
## Se ha de especificar el nombre (ruta relativa) al correr el script, solo esta testeado que funcione correctamente con la herramienta copiar de esa página concreta

echo "Como se llama el dump de CVEdetails?\n\n"
read filename
echo "Proceder con $filename\n\n"

current_dir=$(pwd)
filepath=$current_dir/$filename
echo $filepath

## Genera la tabla completa con CVE | CVSS | Enlace NIST
awk -F'\t' 'NR>1 && $1 ~ /^CVE-/ {
    # Crear enlace NIST
    nist_link = "https://nvd.nist.gov/vuln/detail/" tolower($1)
    print "|" $1 "|" $4 "|" nist_link "|"
}' $filepath > table_CVEs_completa

echo "Tabla generada en: table_CVEs_completa"
