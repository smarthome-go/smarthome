<script lang="ts">
    import type { hardwareNode } from "../types";
    import HWNode from "./HWNode.svelte";

    export let hardwareNodes: hardwareNode[];
</script>

{#each hardwareNodes as n (n.url)}
    <HWNode bind:data={n} />
{/each}
