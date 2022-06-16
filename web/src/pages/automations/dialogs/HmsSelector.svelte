<script lang="ts">
    import { Icon } from '@smui/button'
    import List,{
    Graphic,
    Item,
    PrimaryText,
    SecondaryText,
    Text
    } from '@smui/list'
    import { homescripts } from '../main'

    export let selection: string

    // Checks if the selection is empty for handling preset values as well as no preset values
    $: if (
        $homescripts !== undefined &&
        $homescripts.length > 0 &&
        selection === ''
    )
        selection = $homescripts[0].data.id

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
                selected={selection === item.data.id}
            >
                <Graphic>
                    <Icon class="material-icons">
                        {$homescripts.find((h) => h.data.id === item.data.id)
                            .data.mdIcon}
                    </Icon>
                </Graphic>
                <Text>
                    <PrimaryText>{item.data.name != "" ? item.data.name : "Unknown Name"}</PrimaryText>
                    <SecondaryText>
                        {item.data.description != "" ? item.data.description : "No description provided"}
                    </SecondaryText>
                </Text>
            </Item>
        {/each}
    </List>
</div>

<style lang="scss">
    .list {
        width: 100%;
        height: 100%;
        overflow: auto;
        border-radius: 0.4rem;
        padding: 1rem 0;
    }
</style>
