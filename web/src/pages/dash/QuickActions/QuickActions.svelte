<script lang="ts">
    import Box from "../Box.svelte";
    import { createSnackbar } from "../../../global";
    import {
        getRunningJobs,
        type homescriptJob,
        type homescriptWithArgs,
    } from "../../../homescript";
    import { onMount } from "svelte";
    import Action from "./Action.svelte";
    import Button from "@smui/button/src/Button.svelte";
    import { Label } from "@smui/button";

    let loading = false;

    export let homescripts: homescriptWithArgs[] = [];
    let actions: homescriptWithArgs[] = [];
    let homescriptLoaded = false;

    let jobs: homescriptJob[] = [];
    //let jobsLoaded = false;

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
            // `homescripts` will include all Homescripts
            homescripts = res
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

    // Fetches the current Homescript jobs from the server
    async function loadJobs() {
        loading = true;
        try {
            jobs = await getRunningJobs();
            running = jobs.length;
        } catch (err) {
            $createSnackbar(`Failed to load current Homescript jobs: ${err}`);
        }
        loading = false;
    }

    onMount(() => {
        loadHomescripts().then(loadJobs);
    });
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
            <div class="actions__empty">
                <span class="actions__empty__title">No Quick Actions</span>
                <span class="text-hint">
                    There are currently no Homescript quick-actions.
                </span>
                <Button variant="outlined" href="/homescript">
                    <Label>Create</Label>
                </Button>
            </div>
        {:else}
            {#each actions as data}
                <Action
                    bind:data
                    isAlreadyRunning={jobs.filter(
                        (j) => j.homescriptId === data.data.data.id
                    ).length > 0}
                    on:run={() => running++}
                    on:finish={() => running--}
                />
            {/each}
            <div class="placeholder" />
            <div class="placeholder" />
            <div class="placeholder" />
            <div class="placeholder" />
        {/if}
    </div>
</Box>

<style lang="scss">
    @use "../.././../mixins" as *;
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
        box-sizing: border-box;

        @include mobile {
            justify-content: center;
        }

        &__empty {
            display: flex;
            flex-direction: column;
            align-items: flex-start;

            &__title {
                font-weight: bold;
            }

            .text-hint {
                font-size: 0.9rem;
                margin-bottom: 0.8rem;
            }
        }

        .placeholder {
            @include mobile {
                width: 4.9rem;
                height: 0;
            }
        }
    }
</style>
