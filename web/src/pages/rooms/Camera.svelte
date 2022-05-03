<script lang="ts">
    import IconButton from "@smui/icon-button";
    import Switch from "@smui/switch";
    import { onMount } from "svelte/internal";
    import Progress from "../../components/Progress.svelte";
    import { createSnackbar, hasPermission, sleep } from "../../global";
    import EditSwitch from "./dialogs/switch/EditSwitch.svelte";
    import type { Camera } from "./main";

    export let cameras: Camera[];

    export let id: string;
    export let name: string;
    export let url: string;

    let loading = false;

    let showEditCamera: () => void;

    // Determines if edit button should be shown
    let hasEditPermission: boolean;
    onMount(async () => {
        hasEditPermission = await hasPermission("modifyRooms");
    });
</script>

<div class="switch mdc-elevation--z3">
    <span>{id}:{name}</span>
    <img src="/api/avatar" alt={`video feed of camera: ${id}`} />
    {#if hasEditPermission}
        <IconButton
            class="material-icons"
            title="Edit Switch"
            on:click={showEditCamera}>edit</IconButton
        >
    {/if}
</div>

<style lang="scss">
    @use '../../mixins' as *;
</style>
