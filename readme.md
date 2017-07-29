### Status: Incomplete. The project as it is may not work at all for its intended purpose.

# clonereap
clonereap is a CI tool which responds to pull requests by searching for code
clones in the new codebase. Code clones are bad for all sorts of reasons -
chief among them, it makes your code difficult to maintain at scale.

clonereap depends on its partner project,
[clone-browser](https://github.com/jkobyp/clone-browser) as its frontend.


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
