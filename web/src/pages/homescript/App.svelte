<script lang="ts">
    import Button, { Icon } from "@smui/button";
    import IconButton from "@smui/icon-button";
    import { Label } from "@smui/list";
    import { onMount } from "svelte";
    import Progress from "../../components/Progress.svelte";
    import { createSnackbar, data as userData } from "../../global";
    import Page from "../../Page.svelte";
    import Inputs from "./dialogs/Inputs.svelte";
    import HmsSelector from "./dialogs/HmsSelector.svelte";
    import HomescriptElement from "./HomescriptElement.svelte";
    import { hmsLoaded, homescripts, loading } from "./main";

    let selection = "";
    let addOpen = false;
    let overviewOpen = false;

    // Fetches the available homescripts for the selection and naming
    async function loadHomescripts() {
        $loading = true;
        try {
            let res = await (
                await fetch("/api/homescript/list/personal")
            ).json();

            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            homescripts.set(res); // Move the fetched homescripts into the store
            hmsLoaded.set(true); // Signal that the homescripts are loaded
        } catch (err) {
            $createSnackbar(`Could not load Homescript: ${err}`);
        }
        $loading = false;
    }

    onMount(() => {
        loadHomescripts();
    }); // Load homescripts as soon as the component is mounted
</script>

<Page>
    <div id="header" class="mdc-elevation--z4">
        <h6>Homescript</h6>
        <div>
            <IconButton
                title="Refresh"
                class="material-icons"
                on:click={async () => {
                    await loadHomescripts();
                }}>refresh</IconButton
            >
            {#if $homescripts.length > 0}
                <Button on:click={() => (addOpen = true)}>
                    <Label>Create New</Label>
                    <Icon class="material-icons">add</Icon>
                </Button>
            {/if}
        </div>
    </div>
    <Progress id="loader" bind:loading={$loading} />

    <div id="content">
        <div id="container" class="mdc-elevation--z1">
            <div class="homescripts" class:empty={$homescripts.length == 0}>
                {#if $homescripts.length == 0}
                    <i class="material-icons" id="no-homescripts-icon"
                        >code_off</i
                    >
                    <h6 class="text-hint">No Homescripts</h6>
                    <Button
                        on:click={() => (addOpen = true)}
                        variant="outlined"
                    >
                        <Label>Create New</Label>
                        <Icon class="material-icons">add</Icon>
                    </Button>
                {:else}
                    <HmsSelector bind:selection />
                {/if}
            </div>
        </div>
        <div id="add" class="mdc-elevation--z1">
            <div class="header">
                <h6>Current Homescript</h6>
            </div>
            {#if $homescripts !== undefined && selection !== ''}
            <Inputs
                bind:data={$homescripts[$homescripts.findIndex(h => h.data.id === selection)].data}
            />
        {/if}
        </div>
    </div>
</Page>

<style lang="scss">
    @use "../../mixins" as *;
    #header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 0.1rem 1.3rem;
        box-sizing: border-box;
        background-color: var(--clr-height-1-4);

        h6 {
            margin: .5em 0;

            @include mobile {
                // Hide title on mobile due to space limitations
                display: none;
            }
        }
    }

    #no-homescripts-icon {
        font-size: 5rem;
        color: var(--clr-text-disabled);
    }

    #content {
        display: flex;
        flex-direction: column;
        margin: 1rem 1.5rem;
        gap: 1rem;
        transition-property: height;
        transition-duration: 0.3s;

        @include widescreen {
            flex-direction: row;
            gap: 2rem;
        }
    }

    #container {
        background-color: var(--clr-height-0-1);
        border-radius: 0.4rem;

        @include widescreen {
            width: 50%;
        }
    }

    #add {
        background-color: var(--clr-height-0-1);
        border-radius: 0.4rem;
        padding: 1.5rem;

        h6 {
            margin: .5rem 0;
        }

        @include widescreen {
            width: 50%;
        }
    }
</style>
