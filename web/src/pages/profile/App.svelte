<script lang="ts">
    import IconButton from "@smui/icon-button";
    import Progress from "../../components/Progress.svelte";
    import Page from "../../Page.svelte";
    import Inputs from "./Inputs.svelte";
    import Miscellaneous from "./Miscellaneous.svelte";
    import Permissions from "./dialogs/Permissions.svelte";

    let permissionsOpen = false;
</script>

<Permissions bind:open={permissionsOpen} />
<Page>
    <div id="header" class="mdc-elevation--z4">
        <h6>Your Profile</h6>
        <div id="header__buttons">
            <IconButton title="Refresh" class="material-icons"
                >refresh</IconButton
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
        <div id="miscellaneous" class="mdc-elevation--z1">
            <Miscellaneous />
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
        height: calc(100vh - 60px);
        padding: 1rem 1.5rem;
        box-sizing: border-box;
        flex-direction: column;
        display: flex;
        gap: 1rem;

        @include widescreen {
            flex-direction: row;
        }

        @include mobile {
            min-height: calc(100vh - 48px - 3.5rem);
        }

        #inputs,
        #miscellaneous {
            background-color: var(--clr-height-0-1);
            border-radius: 0.4rem;
            height: 50%;
            width: 100%;

            @include widescreen {
                height: 100%;
                width: 50%;
            }
        }
    }
</style>
