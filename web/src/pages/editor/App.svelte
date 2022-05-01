  
  <script>
    import Monaco from './monaco.svelte'
    
    let monaco;
    window['monaco'] = monaco;
  </script>


<div id="app">
    <Monaco bind:this={monaco} />
  </div>
  
  <style>
    #app {
      width: 100vw;
      height: 100vh;
      position: relative;
      display: grid;
      justify-content: center;
      align-content: center;
      grid-template-columns: 100%;
      grid-template-rows: 100%;
    }
    * :global(.buttons) {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(200px,1fr));
      width: 500px;
    }
  </style>