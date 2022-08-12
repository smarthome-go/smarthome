<script lang="ts">
    import Nodes from "./nodes/Nodes.svelte";
    import type { hardwareNode } from "./types";
    import { createSnackbar } from "../../../global";
    import { onMount } from "svelte";
    import Progress from "../../../components/Progress.svelte";
    import { Icon } from "@smui/button";
    import Fab from "@smui/fab";
    import CreateNode from "./nodes/CreateNode.svelte";
    import IconButton from "@smui/icon-button";

    // Specifies whether the loading indicator should be shown or hidden
    let loading = true;

    /*
        Hardware Nodes
        https://github.com/smarthome-go/node
    */

    // Specifies whether the add hardware node dialog should be open or closed
    let createHardwareNodeOpen = false;

    // Contains all hardware nodes
    let hardwareNodes: hardwareNode[] = [];
    let hardwareNodesLoaded = false;

    //   If the healthcheck should be used, this request will take significantly more time to complete (recommended for manual reloading)
    async function fetchHardwareNodes(withHealthCheck: boolean) {
        loading = true;
        try {
            const res = await (
                await fetch(
                    `/api/system/hardware/node/${
                        withHealthCheck ? "check" : "list"
                    }`
                )
            ).json();
            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            hardwareNodes = res;
            hardwareNodesLoaded = true;
        } catch (err) {
            $createSnackbar(`Failed to load hardware nodes: ${err}`);
        }
        loading = false;
    }

    // Creates a new hardware node
    async function createHardwareNode(
        url: string,
        name: string,
        token: string,
        enabled: boolean
    ) {
        loading = true;
        try {
            const res = await (
                await fetch("/api/system/hardware/node/add", {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        url,
                        data: { name, token, enabled },
                    }),
                })
            ).json();
            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            hardwareNodes = [
                ...hardwareNodes,
                {
                    url,
                    name,
                    token,
                    enabled,
                    online: false,
                },
            ];
            // Reload the nodes in order to obtain the new power state
            setTimeout(() => {
                fetchHardwareNodes(true);
            }, 10);
        } catch (err) {
            $createSnackbar(`Failed to create hardware node: ${err}`);
        }
        loading = false;
    }

    // As soon as the component is mounted, fetch the hardware nodes (without the healthcheck turned on)
    // TODO: allow the user to change this setting per-device (room settings for reference)
    onMount(() => fetchHardwareNodes(false));
</script>

<CreateNode
    bind:open={createHardwareNodeOpen}
    on:create={(e) =>
        createHardwareNode(
            e.detail.url,
            e.detail.name,
            e.detail.token,
            e.detail.enabled
        )}
/>

<div class="hardware">
    <Progress bind:loading />
    <h6>Hardware</h6>

    <!-->Hardware Nodes</-->
    <div class="hardware__type">
        <!-->vendor label starts here</-->
        <div class="hardware__type__label">
            <a
                class="hardware__type__label__name"
                href="https://github.com/smarthome-go/node"
                rel="noopener noreferrer nofollow"
                target="_blank"
                >Nodes
            </a>
            <i class="hardware__type__label__icon material-icons">memory</i>
            <div class="hardware__type__label__right">
                <IconButton
                    disabled={loading}
                    class="material-icons"
                    on:click={() => fetchHardwareNodes(false)}
                    title="Refresh">refresh</IconButton
                >
                <IconButton
                    disabled={loading ||
                        (hardwareNodes.length === 0 && hardwareNodesLoaded)}
                    class="material-icons"
                    on:click={() => fetchHardwareNodes(true)}
                    title="Run Healthcheck">monitor_heart</IconButton
                >
                <Fab
                    color="primary"
                    mini
                    title="Add Node"
                    on:click={() => (createHardwareNodeOpen = true)}
                >
                    <Icon class="material-icons">add</Icon>
                </Fab>
            </div>
        </div>

        <!-->vendor starts here</-->
        <div class="hardware__nodes">
            {#if hardwareNodes.length === 0 && hardwareNodesLoaded}
                <i class="material-icons text-disabled">dns</i>
                <span class="text-hint"> No hardware nodes </span>
            {:else}
                <Nodes bind:hardwareNodes />
            {/if}
        </div>
    </div>

    <!-->Future hardware will be added here</-->
</div>

<style lang="scss">
    // Main list which contains different kinds of manufacturers
    .hardware {
        padding: 1rem 1.5rem;

        h6 {
            margin: 0;
            font-size: 1.1rem;
            color: var(--clr-text-hint);
        }

        &__type {
            &__label {
                display: flex;
                align-items: center;
                gap: 0.4rem;
                margin-top: 1rem;
                margin-bottom: 0.5rem;

                // Any HTML element which can be used to label the coming hardware section
                // Often an `a-tag` which links to a reference page
                &__name {
                    color: var(--clr-text-hint);
                }
                // `i-tag` which contains a MD icon
                &__icon {
                    color: var(--clr-text-hint);
                    font-size: 1.5rem;
                }
                &__right {
                    margin-left: auto;
                    display: flex;
                    align-items: center;
                    gap: 0.5rem;
                }
            }
        }

        /*
           Vendor-specific styles start here
        */

        // The default hardware node device type
        &__nodes {
            display: flex;
            flex-wrap: wrap;
            gap: 1rem;
        }
    }
</style>
