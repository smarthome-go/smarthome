<script lang="ts">
    import type { KitchenComponentDev } from '@smui/snackbar/kitchen'
    import Kitchen,{ ConfigAction } from '@smui/snackbar/kitchen'
    import NavBar from './components/NavBar.svelte'
    import { createSnackbar,data } from './global'

    function contrast(color: string): 'black' | 'white' {
        const r = parseInt(color.slice(1, 3), 16)
        const g = parseInt(color.slice(3, 5), 16)
        const b = parseInt(color.slice(5, 7), 16)
        const a = [r, g, b].map(v => {
            v /= 255
            return v <= 0.03928
                ? v / 12.92
                : Math.pow((v + 0.055) / 1.055, 2.4)
        })
        const luminance = a[0] * 0.2126 + a[1] * 0.7152 + a[2] * 0.0722
        const [darker, brighter] = [1.05, luminance + 0.05].sort()
        return brighter / darker <= 4.5 ? 'black' : 'white'
    }

    $: document.documentElement.classList.toggle('light-theme', !$data.userData.user.darkTheme)
    $: if ($data.loaded) {
        document.documentElement.style.setProperty('--clr-primary-dark', $data.userData.user.primaryColorDark)
        document.documentElement.style.setProperty('--clr-primary-light', $data.userData.user.primaryColorLight)
        document.documentElement.style.setProperty(
            '--clr-on-primary-dark',
            contrast($data.userData.user.primaryColorDark) === 'black' ? '#121212' : '#ffffff',
        )
        document.documentElement.style.setProperty(
            '--clr-on-primary-light',
            contrast($data.userData.user.primaryColorLight) === 'black' ? '#121212' : '#ffffff',
        )
    }

    let kitchen: KitchenComponentDev
    $createSnackbar = (message: string, actions?: ConfigAction[]) => {
        kitchen.push({
            label: message,
            dismissButton: true,
            actions,
        })
    }
</script>

<svelte:head>
    {#if !$data.userData.user.darkTheme}
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
