<script lang="ts">
    import List,{
    Item,
    Meta,
    PrimaryText,
    SecondaryText,
    Text
    } from '@smui/list'
    import { homescripts } from './main'

    export let selection: string

    $: if ($homescripts !== undefined && $homescripts.length > 0)
        $homescripts[0].data.name

    let selectionIndex: number | undefined = undefined
</script>

<div class="list">
    <List
        twoLine
        avatarList
        singleSelection
        bind:selectedIndex={selectionIndex}
    >
        {#each $homescripts as item}
            <Item
                on:SMUI:action={() => (selection = item.data.id)}
                disabled={!item.data.schedulerEnabled}
                selected={selection === item.data.name}
            >
                <!-- <Graphic
                    style="background-image: url(https://place-hold.it/40x40?text={item.data.name
                        .split(' ')
                        .map((val) => val.substring(0, 1))
                        .join('')}&fontsize=16);"
                /> -->
                <Text>
                    <PrimaryText>{item.data.name}</PrimaryText>
                    <SecondaryText>{item.data.description}</SecondaryText>
                </Text>
                <Meta class="material-icons">
                    {#if item.data.schedulerEnabled}
                        {$homescripts.find((h) => h.data.id === item.data.id)
                            .data.mdIcon}
                    {/if}
                </Meta>
            </Item>
        {/each}
    </List>
</div>

<style>
    .list {
        width: 100%;
        height: 80%;
        overflow: auto;
        background-color: var(--clr-height-0-3);
        border-radius: .4rem;
        padding: 1rem 0;
    }
</style>
