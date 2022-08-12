<script lang="ts">
    import IconButton from "@smui/icon-button";
    import Progress from "../../components/Progress.svelte";
    import Page from "../../Page.svelte";
    import Inputs from "./Inputs.svelte";
    import Security from "./Security.svelte";
    import Permissions from "./dialogs/Permissions.svelte";
    import { createSnackbar, data } from "../../global";

    // Specify whether the permissions dialog should be open or closed
    let permissionsOpen = false;

    let forceLoadTokens = false;
    let forceLoadPermissions = false;

    async function fetchUserData() {
        try {
            const res = await (await fetch("/api/user/data")).json();
            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            $data.userData = res;
        } catch (err) {
            $createSnackbar(`Could not fetch user data: ${err}`);
        }
    }
</script>

<Permissions bind:open={permissionsOpen} bind:forceLoadPermissions />

<Page>
    <div id="header" class="mdc-elevation--z4">
        <h6>Your Profile</h6>
        <div id="header__buttons">
            <IconButton
                title="Refresh"
                class="material-icons"
                disabled={forceLoadPermissions || forceLoadTokens}
                on:click={() => {
                    fetchUserData();
                    forceLoadTokens = true;
                }}>refresh</IconButton
            >
            <IconButton
                title="Permissions"
                class="material-icons"
                on:click={() => (permissionsOpen = true)}
                >lock_person</IconButton
            >
        </div>
    </div>
    <Progress id="loader" loading={false} />
    <div id="content">
        <div id="inputs" class="mdc-elevation--z1">
            <Inputs />
        </div>
        <div id="security" class="mdc-elevation--z1">
            <Security bind:forceLoad={forceLoadTokens} />
        </div>
    </div></Page
>

<style lang="scss">
    @use "../../mixins" as *;

    #header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 0.1rem 1.3rem;
        box-sizing: border-box;
        background-color: var(--clr-height-1-4);
        min-height: 3.5rem;

        &__buttons {
            display: flex;
            align-items: center;
        }

        h6 {
            margin: 0.5em 0;
            @include mobile {
                // Hide title on mobile due to space limitations
                display: none;
            }
        }
    }

    #content {
        padding: 1rem 1.5rem;
        box-sizing: border-box;
        flex-direction: column;
        display: flex;
        gap: 1rem;

        @include widescreen {
            height: calc(100vh - 60px);
            flex-direction: row;
        }

        @include mobile {
            min-height: calc(100vh - 48px - 3.5rem);
            padding: 1rem;
        }

        #inputs {
            background-color: var(--clr-height-0-1);
            border-radius: 0.4rem;
            height: 75%;
            width: 100%;

            @include widescreen {
                height: 100%;
                width: 60%;
            }
        }

        #security {
            background-color: var(--clr-height-0-1);
            border-radius: 0.4rem;
            height: 25%;
            width: 100%;

            @include widescreen {
                height: 100%;
                width: 40%;
            }
        }
    }
</style>
