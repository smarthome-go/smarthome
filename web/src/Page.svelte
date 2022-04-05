<script lang="ts">
    import Kitchen from '@smui/snackbar/kitchen'
    import type { KitchenComponentDev } from '@smui/snackbar/kitchen'
    import NavBar from './components/NavBar.svelte'
    import { data, createSnackbar } from './global'

    $: document.documentElement.classList.toggle('light-theme', !$data.userData.darkTheme)

    let kitchen: KitchenComponentDev
    $createSnackbar = (message: string) => {
        kitchen.push({
            label: message,
            dismissButton: true,
        })
    }
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
<Kitchen bind:this={kitchen} dismiss$class="material-icons" />

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
