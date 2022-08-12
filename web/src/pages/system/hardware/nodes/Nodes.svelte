<script lang="ts">
    import type { hardwareNode } from "../types";

    export let hardwareNodes: hardwareNode[];
</script>

{#each hardwareNodes as n (n.url)}
    <span>{n.name}</span>
    <br />
    <span>{n.url}</span>
    <br />
{/each}
