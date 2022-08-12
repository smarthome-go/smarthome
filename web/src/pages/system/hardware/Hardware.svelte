<script lang="ts">
    import Nodes from "./nodes/Nodes.svelte";
    import type { hardwareNode } from "./types";
    import { createSnackbar } from "../../../global";
    import { onMount } from "svelte";
    import Progress from "../../../components/Progress.svelte";

    // Specifies whether the loading indicator should be shown or hidden
    let loading = true;

    /*
        Hardware Nodes
        https://github.com/smarthome-go/node
    */

    // Contains all hardware nodes
    let hardwareNodes: hardwareNode[] = [];

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
        } catch (err) {
            $createSnackbar(`Failed to load hardware nodes: ${err}`);
        }
        loading = false;
    }

    // As soon as the component is mounted, fetch the hardware nodes (without the healthcheck turned on)
    // TODO: allow the user to change this setting per-device (room settings for reference)
    onMount(() => fetchHardwareNodes(false));
</script>

<div class="hardware">
    <Progress bind:loading />
    <h6>Hardware</h6>
    <div class="hardware__nodes">
        <Nodes bind:hardwareNodes />
    </div>
</div>

<style lang="scss">
    .hardware {
        padding: 1rem 1.5rem;

        h6 {
            margin: 0;
            font-size: 1.1rem;
            color: var(--clr-text-hint);
        }
    }
</style>
