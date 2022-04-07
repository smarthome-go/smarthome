<script lang="ts">
    import Button, { Icon, Label } from '@smui/button'
    import Textfield from '@smui/textfield'
    import HelperText from '@smui/textfield/helper-text'
    import LinearProgress from '@smui/linear-progress'
    import type { LinearProgressComponentDev } from '@smui/linear-progress'
    import Snackbar, { Actions } from '@smui/snackbar'
    import type { SnackbarComponentDev } from '@smui/snackbar'
    import IconButton from '@smui/icon-button'
    import Logo from '../../assets/logo.webp'

    let loader: LinearProgressComponentDev

    let snackbar: SnackbarComponentDev
    let errorMessage = ''

    let username = ''
    let password = ''
    let userInvalid = false
    let passwordInvalid = false
    let userDirty = false
    let passwordDirty = false

    $: userInvalid = userDirty && username === ''
    $: passwordInvalid = passwordDirty && password === ''

    let theme = window.localStorage.getItem('theme')
    if (theme === null) theme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
    let darkTheme = theme !== 'light'
    $: document.documentElement.classList.toggle('light-theme', !darkTheme)
    $: window.localStorage.setItem('theme', darkTheme ? 'dark' : 'light')

    async function login(event: SubmitEvent) {
        event.preventDefault()
        if (username === '') userInvalid = true
        if (password === '') passwordInvalid = true
        if (userInvalid || passwordInvalid) return

        loader.getElement().style.opacity = '1'
        const res = await fetch('/api/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password })
        })
        loader.getElement().style.opacity = '0'
        if (res.status === 204) {
            window.location.href = '/'
        } else if (res.status === 401) {
            errorMessage = 'Invalid username and/or password'
            snackbar.open()
        } else {
            errorMessage = 'An unknown error occured. Please try again'
            snackbar.open()
        }
    }
</script>

<svelte:head>
    {#if !darkTheme}
        <link rel="stylesheet" href="/assets/theme-light.css">
    {/if}
</svelte:head>
<main>
    <div id="left" class="mdc-elevation--z8">
        <img src={Logo} alt="logo" />
        <h4>Smarthome</h4>
        <p class="text-hint">Please login to continue</p>
        <svg
            xmlns="http://www.w3.org/2000/svg"
            id="hexagon"
            width="100"
            height="100"
            viewBox="0 0 100 100"
        >
            <linearGradient
                id="gradient"
                gradientTransform="rotate(-45 0.5 0.5)"
            >
                <stop offset="0%" stop-color="var(--clr-primary)" />
                <stop offset="100%" stop-color="var(--clr-height-0-4)" />
            </linearGradient>
            <polygon
                points="25,5 75,5 100,50 75,95 25,95 0,50"
                fill="url(#gradient)"
            />
        </svg>
    </div>
    <div id="right" class="mdc-elevation--z2">
        <LinearProgress id="loader" bind:this={loader} indeterminate />
        <IconButton
            id="theme-toggle"
            on:click={() => darkTheme = !darkTheme}
            class="material-icons"
            title="Toggle light/dark theme"
        >{darkTheme ? 'light_mode' : 'dark_mode'}</IconButton>
        <form on:submit={login}>
            <div>
                <Textfield
                    bind:invalid={userInvalid}
                    bind:dirty={userDirty}
                    bind:value={username}
                    label="Username"
                    variant="outlined"
                >
                    <HelperText validationMsg slot="helper"
                        >This field is required</HelperText
                    >
                </Textfield>
            </div>
            <div>
                <Textfield
                    bind:invalid={passwordInvalid}
                    bind:dirty={passwordDirty}
                    bind:value={password}
                    label="Password"
                    type="password"
                    variant="outlined"
                >
                    <HelperText validationMsg slot="helper"
                        >This field is required</HelperText
                    >
                </Textfield>
            </div>
            <Button variant="raised">
                <Icon class="material-icons">login</Icon>
                <Label>Login</Label>
            </Button>
        </form>
    </div>
    <Snackbar bind:this={snackbar}>
        <Label>{errorMessage}</Label>
        <Actions>
            <IconButton class="material-icons" title="Dismiss">close</IconButton>
        </Actions>
    </Snackbar>
</main>

<style lang="scss">
    @use '../../mixins' as *;

    :global body {
        margin: 1rem;
    }

    main {
        position: fixed;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
        display: flex;
        align-items: center;
        height: 34rem;
        width: min(50rem, 100% - 4rem);

        @include mobile {
            flex-direction: column;
            width: 100%;
            height: auto;
            position: relative;
            top: 0;
            left: 0;
            transform: none;
            padding-inline: 1rem;
            box-sizing: border-box;
            z-index: 0;
        }
    }

    #left {
        border-radius: 0.4rem;
        width: 40%;
        height: calc(100% + 2rem);
        box-sizing: border-box;
        position: relative;
        background-color: var(--clr-height-0-12);

        display: flex;
        flex-direction: column;
        align-items: center;
        padding-top: 3rem;

        img {
            height: 6rem;
            aspect-ratio: 1;
        }
        h4 {
            margin-bottom: 0.3em;
        }

        @include mobile {
            width: calc(100% + 2rem);
            height: auto;
            padding-block: 2rem;
        }
    }
    #hexagon {
        height: 15rem;
        width: 15rem;
        // Mimics z3 elevation
        filter: drop-shadow(0px 1px 3px rgb(0 0 0 / 20%))
            drop-shadow(0px 3px 4px rgb(0 0 0 / 14%))
            drop-shadow(0px 1px 8px rgb(0 0 0 / 12%));
        position: absolute;
        right: 0;
        bottom: 2rem;
        z-index: -5;
        transform: translateX(50%);

        @include mobile {
            height: 10rem;
            width: 10rem;
            bottom: 0;
            right: 3rem;
            transform: translate(0, 50%) rotate(90deg);
        }
    }
    #right {
        border-radius: 0 0.4rem 0.4rem 0;
        width: 60%;
        height: 100%;
        padding-top: 5rem;
        box-sizing: border-box;
        z-index: -10;
        position: relative;
        overflow: hidden;
        background-color: var(--clr-height-0-2);

        form {
            display: flex;
            flex-direction: column;
            align-items: center;
            gap: 1rem;

            @include mobile {
                & > div {
                    max-width: calc(100% - 2rem);
                }
            }
        }

        @include mobile {
            width: 100%;
            height: auto;
            border-radius: 0 0 0.4rem 0.4rem;
            padding-top: 7rem;
            padding-bottom: 3rem;
        }
    }
    main :global #loader {
        position: absolute;
        top: 0;
        opacity: 0;
    }
    main :global #theme-toggle {
        position: absolute;
        top: 1rem;
        right: 2rem;

        @include mobile {
            right: auto;
            left: 2rem;
        }
    }
</style>
