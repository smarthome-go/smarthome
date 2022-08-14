<script lang="ts">
    import Box from "../Box.svelte";
    import { createSnackbar } from "../../../global";
    import type { homescriptWithArgs } from "../../../homescript";
    import { onMount } from "svelte";
    import Action from "./Action.svelte";

    let loading = false;

    let actions: homescriptWithArgs[] = [];
    let homescriptLoaded = false;

    let running = 0;

    // Fetches the available Homescripts for displaying the quick actions
    async function loadHomescripts() {
        loading = true;
        try {
            let res = await (
                await fetch("/api/homescript/list/personal/complete")
            ).json();
            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            // Just include Homescript which have enabled quick actions
            actions = res.filter(
                (h: homescriptWithArgs) => h.data.data.quickActionsEnabled
            );
            // Signal that the actions have been successfully loaded
            homescriptLoaded = true;
            // Create the boolean list
        } catch (err) {
            $createSnackbar(`Could not load Homescript Quick Actions: ${err}`);
        }
        loading = false;
    }

    onMount(loadHomescripts);
</script>

<Box bind:loading>
    <a href="/homescript" slot="header" class="title">Quick Actions</a>
    <span slot="header-right" class="job-count">
        {#if running === 0}
            Idle
        {:else}
            {running} Job{running !== 1 ? "s" : ""} running
        {/if}
    </span>
    <div class="actions" slot="content">
        {#if homescriptLoaded && actions.length === 0}
            No Actions create button here
        {:else}
            {#each actions as data}
                <Action
                    bind:data
                    on:run={() => running++}
                    on:finish={() => running--}
                />
            {/each}
        {/if}
    </div>
</Box>

<style lang="scss">
    .title {
        color: var(--clr-primary-hint);
        font-weight: bold;
        text-decoration: none;
    }
    .job-count {
        color: var(--clr-text-hint);
        font-size: 0.8rem;
    }
    .actions {
        display: flex;
        gap: 0.5rem;
        flex-wrap: wrap;
        align-content: flex-start;
        height: 100%;
        padding: 0.4rem 0.125rem;
    }
</style>
