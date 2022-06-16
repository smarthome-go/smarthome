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
        selection = $homescripts[0].data.data.id

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
                on:SMUI:action={() => (selection = item.data.data.id)}
                selected={selection === item.data.data.id}
            >
                <Graphic>
                    <Icon class="material-icons">
                        {$homescripts.find((h) => h.data.data.id === item.data.data.id)
                            .data.data.mdIcon}
                    </Icon>
                </Graphic>
                <Text>
                    <PrimaryText>{item.data.data.name != "" ? item.data.data.name : "Unknown Name"}</PrimaryText>
                    <SecondaryText>
                        {item.data.data.description != "" ? item.data.data.description : "No description provided"}
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
        border-radius: 0.4rem;
        // padding: 1rem 0;
    }
</style>
