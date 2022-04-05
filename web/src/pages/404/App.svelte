<script lang="ts">
    import Button, { Label } from '@smui/button'
    import IconButton from '@smui/icon-button'

    let theme = window.localStorage.getItem('theme')
    if (theme === null) theme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
    let darkTheme = theme !== 'light'
    $: document.documentElement.classList.toggle('light-theme', !darkTheme)
    $: window.localStorage.setItem('theme', darkTheme ? 'dark' : 'light')
</script>

<svelte:head>
    {#if !darkTheme}
        <link rel="stylesheet" href="/assets/theme-light.css">
    {/if}
</svelte:head>
<main class="mdc-elevation--z2">
    <IconButton
        id="theme-toggle"
        on:click={() => darkTheme = !darkTheme}
        class="material-icons"
        title="Toggle light/dark theme"
    >{darkTheme ? 'light_mode' : 'dark_mode'}</IconButton>
    <h1>404</h1>
    <h4>Page not found</h4>
    <Button variant="raised" href="/dash">
        <Label>Go back</Label>
    </Button>
</main>

<style lang="scss">
    main {
        background-color: var(--clr-height-0-2);
        border-radius: .4rem;
        position: fixed;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
        padding: 3rem;
        // min-width: 50rem;

        h1 { margin-bottom: 0; }
    }

    main :global #theme-toggle {
        position: absolute;
        top: 1rem;
        right: 2rem;
    }
</style>
