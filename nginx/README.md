# Nginx configuration

I'd recommend running the emote server behind nginx to use your self-signed certificates.

In `app.example.conf`:
 - Change the two `proxy_pass` values to the appropriate address for your server
 - Change the two `proxy_set_header Host` values to the appropriate hosts
 - Configure SSL certificates
