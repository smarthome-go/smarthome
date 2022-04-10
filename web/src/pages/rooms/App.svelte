<script lang="ts">
    import IconButton from '@smui/icon-button';
    import Tab,{ Label } from '@smui/tab';
    import TabBar from '@smui/tab-bar';
    import Progress from '../../components/Progress.svelte';
    import { createSnackbar,sleep } from '../../global';
    import Page from '../../Page.svelte';
    import Switch from './Switch.svelte';

    interface RoomResponse {
        id: string,
        name: string,
        description: string,
        switches: SwitchResponse[],
    }
    interface SwitchResponse {
        id: string,
        name: string,
        powerOn: boolean,
        watts: number,
    }

    let loading = false
    let rooms: RoomResponse[]
    let currentRoom: RoomResponse
    $: if (currentRoom !== undefined) window.localStorage.setItem('current_room', currentRoom.id)

    async function loadRooms(updateExisting: boolean = false) {
        loading = true
        try {
            const res = await (await fetch('/api/room/list/personal')).json()
            if (res.success === false) throw Error()
            if (updateExisting) {
                for (const room of rooms) {
                    room.switches = (res as RoomResponse[]).find(r => r.id === room.id).switches
                }
            } else rooms = res
            const roomId = window.localStorage.getItem('current_room')
            const room = roomId === null ? undefined : rooms.find(r => r.id === roomId)
            currentRoom = room === undefined ? rooms[0] : room
        } catch {
            $createSnackbar('Could not load rooms', [{
                onClick: () => loadRooms(updateExisting),
                text: 'retry',
            }])
        }
        loading = false
        while (rooms === undefined) await sleep(10)
    }
</script>

<Page>
    <div id="tabs" class="mdc-elevation--z8">
        {#await loadRooms() then}
        <TabBar tabs={rooms} let:tab={room} bind:active={currentRoom} key={tab => tab.id}>
            <Tab tab={room} minWidth>
                <Label>{room.name}</Label>
            </Tab>
        </TabBar>
        {/await}
        <IconButton class="material-icons" on:click={() => loadRooms(true)}>refresh</IconButton>
        <Progress id="loader" bind:loading />
    </div>
    <div id="content">
        <div id="switches" class="mdc-elevation--z1">
            {#each currentRoom !== undefined ? currentRoom.switches : [] as sw (sw.id)}
                <Switch bind:checked={sw.powerOn} id={sw.id} label={sw.name} />
            {/each}
        </div>
        <div id="cameras" class="mdc-elevation--z1"></div>
    </div>
</Page>

<style lang="scss">
    @use '../../mixins' as *;

    #tabs {
        background-color: var(--clr-height-0-8);
        min-height: 48px;
        position: relative;
        display: flex;

        & :global(#loader) {
            position: absolute;
            inset: 0;
            top: auto;
        }
    }

    #content {
        min-height: calc(100vh - 48px);
        padding: 1rem 1.5rem;
        display: flex;
        gap: 1rem;
        flex-direction: column;
        box-sizing: border-box;

        @include widescreen { flex-direction: row; }
        @include mobile { min-height: calc(100vh - 48px - 3.5rem); }
    }
    #switches {
        background-color: var(--clr-height-0-1);
        padding: 1.5rem;
        border-radius: .4rem;
        display: flex;
        flex-wrap: wrap;
        gap: 1rem;
        align-content: flex-start;
        box-sizing: border-box;
        min-height: calc(100% - 16rem);
        flex-grow: 1;

        @include widescreen { min-height: 100%; }
    }
    #cameras {
        background-color: var(--clr-height-0-1);
        height: 15rem;
        border-radius: .4rem;
        padding: 1.5rem;
        box-sizing: border-box;

        @include widescreen {
            height: auto;
            min-height: 100%;
            width: 20rem;
        }
    }
</style>
