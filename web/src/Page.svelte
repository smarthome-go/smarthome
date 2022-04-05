<script lang="ts">
    import Snackbar, { Label, Actions } from '@smui/snackbar'
    import IconButton from '@smui/icon-button'
    import NavBar from './components/NavBar.svelte'
    import { data, infoBar } from './global'

    $: document.documentElement.classList.toggle('light-theme', !$data.userData.darkTheme)
</script>

<svelte:head>
    {#if !$data.userData.darkTheme}
        <link rel="stylesheet" href="/assets/theme-light.css">
    {/if}
</svelte:head>
<NavBar />
<main>
    <slot></slot>
</main>
<Snackbar bind:this={$infoBar.bar}>
    <Label>{$infoBar.message}</Label>
    <Actions>
        <IconButton class="material-icons" title="Dismiss">close</IconButton>
    </Actions>
</Snackbar>

<style lang="scss">
    @use './mixins' as *;

    main {
        margin-left: 5.125rem;
        transition: margin-left .3s;
        @include mobile {
            margin-left: 0;
            margin-top: 3.5rem;
        }
        @include widescreen {
            margin-left: 13rem;
        }
    }
</style>
