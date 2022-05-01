<script context="module">
    let monaco_promise;
    let _monaco;
    monaco_promise = import('./monaco.js');
    monaco_promise.then(mod => {
      _monaco = mod.default;
    })
    let destroyed = false
  </script>
  
  <script>
    import { onMount } from 'svelte'
    let monaco;
    let container;
    let editor;
    onMount(() => {
          if (_monaco) {
        monaco = _monaco;

        editor = monaco.editor.create(
          container
        )
              // createEditor(mode || 'svelte').then(() => {
              // 	if (editor) editor.setValue(code || '');
        // });
          } else {
            monaco_promise.then(async mod => {
          monaco = mod.default;
          editor = monaco.editor.create(
            container,
            {
              value: [
                "switch('s1', on)",
                "if weaher == 'rainy' {",
                "\tswitch('s1', off)",
                '}'
              ].join('\n'),
              language: 'homescript'
            }
          )
              });
          }
          return () => {
              destroyed = true;
          }
    });
  </script>
  
  <div class="monaco-container" bind:this={container} style="height: 100; text-align: left">
  </div>