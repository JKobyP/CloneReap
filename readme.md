### Status: Incomplete. The project as it is may not work at all for its intended purpose.

# clonereap
clonereap is a CI tool which responds to pull requests by searching for code
clones in the new codebase. Code clones are bad for all sorts of reasons -
chief among them, it makes your code difficult to maintain at scale.

clonereap is a Go project, but it also serves up React.

clonereap depends on its partner project,
[clone-browser](https://github.com/jkobyp/clone-browser) as its frontend.


### To build:
1. download github.com/jkobyp/clone-browser
2. `npm install` and `npm run build`
3. copy the dist folder to the project directory
4. make sure index.html contains:
```html
...
<body>
    <div id="root"> </div>
    <script src="dist/bundle.js"> </script>
</body>
```

5. build and run the project
