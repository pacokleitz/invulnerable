# Public Assets

This folder contains static assets that will be served directly by the web server.

## Favicons

Place your favicon files here. Recommended files:

- `favicon.ico` - Classic favicon (16x16, 32x32, 48x48)
- `favicon-16x16.png` - 16x16 PNG
- `favicon-32x32.png` - 32x32 PNG
- `apple-touch-icon.png` - 180x180 for iOS
- `android-chrome-192x192.png` - 192x192 for Android
- `android-chrome-512x512.png` - 512x512 for Android

After adding favicon files, update `/index.html` with:

```html
<link rel="icon" type="image/x-icon" href="/favicon.ico">
<link rel="icon" type="image/png" sizes="32x32" href="/favicon-32x32.png">
<link rel="icon" type="image/png" sizes="16x16" href="/favicon-16x16.png">
<link rel="apple-touch-icon" sizes="180x180" href="/apple-touch-icon.png">
```

## Other Static Assets

You can also place other static assets here like:
- `robots.txt`
- `manifest.json` (for PWA)
- Images referenced directly by URL
