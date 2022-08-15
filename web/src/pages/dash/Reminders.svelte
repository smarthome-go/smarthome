<script lang="ts">
    import Box from "./Box.svelte";

    let loading = false;
</script>

<Box bind:loading>
    <span slot="header">Reminders</span>
    <div slot="content">Reminders here</div>
</Box>
