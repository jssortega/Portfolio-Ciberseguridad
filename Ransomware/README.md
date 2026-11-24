# Ransomware

- Autor: Jesús Ortega Castillo

## Aclaración

Este Ransomware fue creado por motivos educacionales y no ha sido usado bajo ningún concepto con intenciones maliciosas, todas las pruebas fueron ejecutadas en una máquina virtual aislada.

## Explicación

Hay 3 programas:

-Ransomware: Primero elimina las copias shadow del equipo, posteriormente cifra los archivos con extensiones indicadas, para el cifrado usa AES con una clave que calcula de forma aleatoria. Una vez cifrados los archivos envía esa clave a otro equipo (el del atacante) que esté escuchando con netcat en el puerto 4444 y elimina esa clave para que no quede rastro de ella. Por último este ransomware también realiza una captura de pantalla al equipo de la víctima cada 5 segundos y envía esas capturas al atacante que estará a la escucha y procesará esas imagenes con el programa receptorImagenes.

-Vacuna: descifra todos los archivos que han sido cifrados con el ransomware, cuando lo ejecutas pide al usuario la clave, una vez introducida descifra esos archivos.

-ReceptorImagenes: este abre el puerto 5555 y lo pone a la escucha para recibir y procesar las capturas que le envía el Ransomware.
