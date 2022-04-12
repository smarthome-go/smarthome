<script lang="ts">
    import Button,{ Label } from '@smui/button'
    import IconButton from '@smui/icon-button'
    import { createSnackbar,data,sleep } from '../global'
    import NotificationCard from './Notification.svelte'
    import Progress from './Progress.svelte'

    export let hidden = true

    let loaded = false
    let loading = false

    $: if (!hidden && !loaded) load().then(() => (loaded = true))
    $: if (loaded) $data.notificationCount = $data.notifications.length

    async function load() {
        loading = true
        try {
            const res = await (
                await fetch('/api/user/notification/list')
            ).json()
            if (res.success === false) throw Error()
            $data.notifications = res
        } catch {
            $createSnackbar('Could not refresh notifications')
        }
        loading = false
    }
    async function deleteAll() {
        loading = true
        try {
            const res = await (
                await fetch('/api/user/notification/delete/all', {
                    method: 'DELETE',
                })
            ).json()
            if (!res.success) throw Error()

            const notifications = [...$data.notifications]
            const viewportHeight = Math.max(
                window.innerHeight,
                document.documentElement.clientHeight
            )
            for (const notification of notifications) {
                const element = document.getElementById(
                    `notification-${notification.id}`
                )
                // Whether element is outside visible area (+500px threshold)
                if (element.getBoundingClientRect().top > viewportHeight + 500)
                    break
                element.style.transform = 'translateX(-110%)'
                await sleep(50)
            }
            await sleep(300)
            $data.notifications = []
        } catch {
            $createSnackbar('Could not delete notifications')
        }
        loading = false
    }
</script>

<div id="drawer" class:hidden class:mdc-elevation--z8={!hidden}>
    <Progress id="loader" bind:loading />
    <div id="header">
        <Button
            on:click={deleteAll}
            disabled={$data.notifications.length === 0}
        >
            <Label>Delete All</Label>
        </Button>
        <IconButton
            on:click={() => load()}
            disabled={loading}
            class="material-icons">refresh</IconButton
        >
    </div>
    <div id="list">
        {#if loaded}
            {#each $data.notifications as notification (notification.id)}
                <NotificationCard {...notification} />
            {/each}
        {:else}
            {#each [...Array($data.notificationCount).keys()] as _}
                <NotificationCard dummy />
            {/each}
        {/if}
    </div>
    {#if $data.notificationCount === 0}
        <div id="done">
            <i class="material-icons">done</i>
            <span>All caught up, no notifications</span>
        </div>
    {/if}
</div>

<style lang="scss">
    @use '../mixins' as *;

    #drawer :global #loader {
        position: absolute;
        top: 0;
        left: 0;
        opacity: 0;
        &.visible {
            opacity: 1;
        }
    }
    #header {
        display: flex;
        justify-content: flex-end;
        align-items: center;
        gap: 0.5rem;
        margin-block: 0.5rem;
    }
    #drawer {
        width: 25rem;
        z-index: -10;
        position: absolute;
        top: 0;
        bottom: 0;
        right: 0;
        transform: translateX(100%);
        transition-property: transform, box-shadow;
        transition-duration: 0.3s;
        padding-inline: 1rem;
        overflow-y: scroll;
        overflow-x: hidden;
        background-color: var(--clr-height-0-8);
        @include mobile {
            top: auto;
            left: 0;
            width: auto;
            height: calc(100vh - 3.5rem);
            box-sizing: border-box;
            transform: translateY(100%);
        }

        &.hidden {
            transform: translateX(0%);
            @include mobile {
                transform: translateY(0%);
            }
        }
    }
    #list {
        display: flex;
        flex-direction: column;
        align-items: center;
    }
    #done {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 1rem;
        animation: fade-in 0.2s /* ease 0 once */;

        i {
            font-size: 4rem;
            margin-top: 1rem;
        }
    }
    @keyframes fade-in {
        0% {
            opacity: 0;
        }
        100% {
            opacity: 1;
        }
    }
</style>
