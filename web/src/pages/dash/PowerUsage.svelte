<script lang="ts">
    import Box from "./Box.svelte";

    export let loading = false
</script>

<Box bind:loading>
    <span slot="header">Power Usage</span>
    <div slot="content">
        <h6>Content Here</h6>
    </div>
</Box>

<style lang="scss">
    h6 {
        margin: 0;
    }
</style>
