To build:
download github.com/jkobyp/clone-browser
npm install and npm run build
copy the dist folder to the project directory
make sure index.html contains:
```html
...
<body>
    <div id="root"> </div>
    <script src="dist/bundle.js"> </script>
</body>
```

build and run the project
