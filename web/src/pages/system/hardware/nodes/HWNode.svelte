<script lang="ts">
    import IconButton, { Icon } from "@smui/icon-button";
    import Progress from "../../../../components/Progress.svelte";
    import type { hardwareNode } from "../types";
    import { createSnackbar } from "../../../../global";
    import { createEventDispatcher } from "svelte";
    import EditNode from "./EditNode.svelte";
    import DeleteNode from "./DeleteNode.svelte";

    // Event dispatcher
    const dispatch = createEventDispatcher();

    // If the edit dialog should be open or closed
    let editOpen = false;

    // If the delete dialog should be open or closed
    let deleteOpen = false;

    // If the loading indicator should be shown or hidden
    let loading = false;

    export let data: hardwareNode = {
        url: "",
        name: "",
        token: "",
        enabled: false,
        online: false,
    };

    // Deletes this hardware node
    async function deleteHardwareNode() {
        loading = true;
        try {
            const res = await (
                await fetch("/api/system/hardware/node/delete", {
                    method: "DELETE",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        url: data.url,
                    }),
                })
            ).json();
            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            // If the request was successful, send a deletion request upstream
            dispatch("delete", null);
        } catch (err) {
            $createSnackbar(`Failed to delete hardware node: ${err}`);
        }
        loading = false;
    }

    // Edits this hardware node. If the request was successful, update the value in the GUI
    async function editHardwareNode(
        name: string,
        token: string,
        enabled: boolean
    ) {
        loading = true;
        try {
            const res = await (
                await fetch("/api/system/hardware/node/modify", {
                    method: "PUT",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        url: data.url,
                        data: {
                            name,
                            token,
                            enabled,
                        },
                    }),
                })
            ).json();
            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            // If the request was successful, update the new values in the GUI
            data.name = name;
            data.token = token;
            data.enabled = enabled;
        } catch (err) {
            $createSnackbar(`Failed to edit hardware node: ${err}`);
        }
        loading = false;
    }
</script>

<EditNode
    {data}
    bind:open={editOpen}
    on:edit={(e) =>
        editHardwareNode(e.detail.name, e.detail.token, e.detail.enabled)}
/>

<DeleteNode bind:open={deleteOpen} on:delete={deleteHardwareNode}>jk</DeleteNode
>

<div class="node mdc-elevation--z3" class:disabled={!data.enabled}>
    <div class="node__header">
        <div class="node__header__top">
            <span class="node__header__top__name">{data.name}</span>
            <i
                class="material-icons node__header__top__online"
                class:offline={!data.online}
                >{data.online ? "lan" : "perm_scan_wifi"}</i
            >
        </div>
        <span class="node__header__url text-hint">{data.url}</span>
    </div>
    {#if !data.online}
        <span class="node__offline">OFFLINE</span>
    {:else}
        <span class="node__online">ONLINE</span>
    {/if}
    <div class="node__footer">
        <Progress bind:loading type="circular" />
        <IconButton
            class="material-icons"
            on:click={() => (editOpen = true)}
            title="Edit">edit</IconButton
        >
        <IconButton
            class="material-icons"
            on:click={() => (deleteOpen = true)}
            title="Delete">delete</IconButton
        >
    </div>
</div>

<style lang="scss">
    @use "../../../../mixins" as *;

    .node {
        background-color: var(--clr-height-1-3);
        position: relative;
        border-radius: 0.3rem;
        padding: 1rem;
        height: 6rem;
        width: 14rem;

        &.disabled {
            opacity: 65%;
        }

        @include widescreen {
            // Adjusted so it fits on 1080p and 3440p widescreens
            width: 16rem;
        }

        &__header {
            &__top {
                display: flex;
                justify-content: space-between;
                align-items: center;

                &__name {
                    display: block;
                    font-weight: bold;
                }

                &__online {
                    font-size: 1.25rem;
                    color: var(--clr-success);

                    &.offline {
                        color: var(--clr-error);
                    }
                }
            }
            &__url {
                font-size: 0.85rem;
            }
        }

        &__offline {
            font-size: 0.75rem;
            color: var(--clr-error);
            font-weight: bold;
        }
        &__online {
            font-size: 0.75rem;
            color: var(--clr-success);
            font-weight: bold;
        }

        &__footer {
            position: absolute;
            right: 10px;
            bottom: 5px;
        }
    }
</style>
