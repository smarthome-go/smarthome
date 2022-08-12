<script lang="ts">
    import type { hardwareNode } from "../types";
    import HWNode from "./HWNode.svelte";

    export let hardwareNodes: hardwareNode[];
</script>

{#each hardwareNodes as n (n.url)}
    <HWNode
        on:delete={() =>
            (hardwareNodes = hardwareNodes.filter((h) => h.url !== n.url))}
        bind:data={n}
    />
{/each}
