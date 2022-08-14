<script lang="ts">
    import Box from "./Box.svelte";

    let loading = false;
</script>

<Box bind:loading>
    <span slot="header">Weather</span>
    <div slot="content">
        Weather is coming soon
    </div>
</Box>
