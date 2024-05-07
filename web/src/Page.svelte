<script lang="ts">
    import type { KitchenComponentDev } from "@smui/snackbar/kitchen";
    import Kitchen from "@smui/snackbar/kitchen";
    import type { ConfigAction } from "@smui/snackbar/kitchen";
    import NavBar from "./components/NavBar.svelte";
    import { contrast, createSnackbar, data } from "./global";
    import { onMount } from "svelte";

    export let persistentSlimNav = false
    let slimNav = persistentSlimNav

    $: document.documentElement.classList.toggle(
        "light-theme",
        !$data.userData.user.darkTheme
    );
    $: if ($data.loaded) {
        document.documentElement.style.setProperty(
            "--clr-primary-dark",
            $data.userData.user.primaryColorDark
        );
        document.documentElement.style.setProperty(
            "--clr-primary-light",
            $data.userData.user.primaryColorLight
        );
        document.documentElement.style.setProperty(
            "--clr-on-primary-dark",
            contrast($data.userData.user.primaryColorDark) === "black"
                ? "#121212"
                : "#ffffff"
        );
        document.documentElement.style.setProperty(
            "--clr-on-primary-light",
            contrast($data.userData.user.primaryColorLight) === "black"
                ? "#121212"
                : "#ffffff"
        );
    }

    let kitchen: KitchenComponentDev;
    $createSnackbar = (message: string, actions?: ConfigAction[]) => {
        kitchen.push({
            label: message,
            dismissButton: true,
            actions,
        });
    };

    onMount(() => {
        slimNav = persistentSlimNav
    })
</script>

<svelte:head>
    {#if !$data.userData.user.darkTheme}
        <link rel="stylesheet" href="/assets/theme-light.css" />
    {/if}
</svelte:head>

<NavBar persistentClose={persistentSlimNav} />

<main class:slimNav={slimNav}>
    <slot />
</main>

<Kitchen bind:this={kitchen} dismiss$class="material-icons" />

<style lang="scss">
    @use "./mixins" as *;

    main {
        margin-left: 5.125rem;
        transition: margin-left 0.3s;

        @include mobile {
            margin-left: 0;
            margin-top: 3.5rem;
        }

        @include widescreen {
            margin-left: 13rem;
        }

        &.slimNav {
            margin-left: 5.125rem;
            transition: margin-left 0.3s;

            @include mobile {
                margin-left: 0;
            }
        }
    }
</style>
