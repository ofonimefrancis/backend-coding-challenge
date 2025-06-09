#!/bin/bash

# Create swagger-ui directory
mkdir -p docs/swagger-ui

# Download Swagger UI
curl -L https://github.com/swagger-api/swagger-ui/archive/refs/tags/v5.11.0.tar.gz -o swagger-ui.tar.gz

# Extract only the dist folder
tar -xzf swagger-ui.tar.gz --strip-components=1 swagger-ui-5.11.0/dist/

# Move files to docs/swagger-ui
mv dist/* docs/swagger-ui/

# Clean up
rm -rf dist swagger-ui.tar.gz

# Create custom index.html
cat > docs/swagger-ui/index.html << 'EOF'
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Swagger UI</title>
  <link rel="stylesheet" type="text/css" href="swagger-ui.css" >
  <link rel="icon" type="image/png" href="favicon-32x32.png" sizes="32x32" />
  <link rel="icon" type="image/png" href="favicon-16x16.png" sizes="16x16" />
  <style>
    html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
    *, *:before, *:after { box-sizing: inherit; }
    body { margin:0; background: #fafafa; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="swagger-ui-bundle.js"></script>
  <script src="swagger-ui-standalone-preset.js"></script>
  <script>
  window.onload = function() {
    window.ui = SwaggerUIBundle({
      url: "/swagger/openapi.yml",
      dom_id: '#swagger-ui',
      deepLinking: true,
      presets: [
        SwaggerUIBundle.presets.apis,
        SwaggerUIStandalonePreset
      ],
      plugins: [
        SwaggerUIBundle.plugins.DownloadUrl
      ],
      layout: "StandaloneLayout"
    });
  }
  </script>
</body>
</html>
EOF

echo "Swagger UI has been set up in docs/swagger-ui/" 