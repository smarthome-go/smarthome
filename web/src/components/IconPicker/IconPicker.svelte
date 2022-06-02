<script lang="ts">
    import Dialog, {
        Title,
        Header,
        Content,
        Actions,
        InitialFocus,
    } from "@smui/dialog";
    import Button, { Icon, Label } from "@smui/button";
    import List, { Item, Graphic, Text } from "@smui/list";
    import Radio from "@smui/radio";
    import { onMount } from "svelte/internal";
    import { createSnackbar } from "../../global";
    import Textfield from "@smui/textfield";
    import HelperText from "@smui/textfield/helper-text";
    import Progress from "../Progress.svelte";

    interface iconObj {
        name: string;
        tags: string[];
    }

    // Other variables
    export let open = false;
    export let selected = "";
    export let title = "Select Icon";
    export let loading = false;
    let loaded = false;

    // All loaded icons
    let rawIcons: iconObj[] = [];

    // Currently shown icons (search / slice)
    let currentIcons: iconObj[] = [];

    let currentIconsTemp: iconObj[] = [];

    // Controls which icons are shown
    export let searchTerm = "";

    // Search every time the search-keywoard changes
    $: doSearch(searchTerm);

    // If the user is done typing the word, display search results
    $: if (searchingCount === 0) currentIcons = currentIconsTemp;

    // Keep track of searching
    let searchingCount = 0;
    let searching = false;
    $: searching = searchingCount !== 0;

    async function doSearch(term: string) {
        searchingCount++;
        currentIconsTemp = rawIcons
            .filter(
                (i) =>
                    i.name === term ||
                    i.name.includes(term) ||
                    i.tags.includes(term)
            )
            .slice(0, 50);
        setTimeout(() => {
            searchingCount--;
        }, 500);
    }

    async function loadIcons() {
        loading = true;
        try {
            const res = await (await fetch("/assets/icons.json")).json();
            rawIcons = res.icons;
            // Perform an empty or preset search based on context of the dialog
            doSearch(searchTerm);
            loaded = true;
        } catch (err) {
            $createSnackbar(`Failed to load icon list: Error: ${err}`);
        }
        loading = false;
    }

    onMount(() => loadIcons().then(() => doSearch(selected)));
</script>

<Dialog
    bind:open
    aria-labelledby="over-fullscreen-confirmation-title"
    aria-describedby="over-fullscreen-confirmation-content"
    slot="over"
    selection
>
    <Header>
        <div class="header">
            <Title id="over-fullscreen-confirmation-title">{title}</Title>
            <Progress bind:loading={searching} type="circular" />
        </div>
    </Header>
    {#if loaded}
        <Content id="over-fullscreen-confirmation-content">
            <div class="search">
                <Textfield
                    style="width: 100%;"
                    helperLine$style="width: 100%;"
                    bind:value={searchTerm}
                    label="Icon Search"
                >
                    <HelperText slot="helper"
                        >Enter tags you want to search for</HelperText
                    >
                </Textfield>
            </div>
            {#if currentIcons.length > 0}
                <div class="list" class:disabled={searching}>
                    <List radioList>
                        {#each currentIcons as ic (ic.name)}
                            <Item use={[InitialFocus]}>
                                <Graphic>
                                    <Radio
                                        bind:group={selected}
                                        value={ic.name}
                                    />
                                </Graphic>
                                <Icon
                                    class="material-icons"
                                    style="padding-right: 1rem;">{ic.name}</Icon
                                >
                                <Text>
                                    <span class="label">
                                        {ic.name}
                                    </span>
                                </Text>
                            </Item>
                        {/each}
                    </List>
                </div>
            {:else}
                <div class="no-res">
                    <i class="material-icons">search_off</i>
                    <h6>No results</h6>
                </div>
            {/if}
        </Content>
    {/if}
    <Actions>
        <Button>
            <Label>Cancel</Label>
        </Button>
        <Button action="accept" disabled={selected === ""}>
            <Label>Select</Label>
        </Button>
    </Actions>
</Dialog>

<style lang="scss">
    @use "../../mixins" as *;
    .search {
        padding: 0 1rem;
    }
    .list {
        height: 40vh;
        width: 18rem;
        overflow-y: scroll;
        transition: 0.000001s opacity linear;

        &.disabled {
            opacity: 50%;
            transition-duration: 0.4s;
        }

        @include mobile {
            height: 100%;
            width: 100%;
        }
    }
    .label {
        color: var(--clr-text-disabled);
    }
    .header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        margin-right: 1.5rem;
    }
    .no-res {
        height: 40vh;
        width: 18rem;
        padding-top: 3rem;
        box-sizing: border-box;
        display: flex;
        align-items: center;
        flex-direction: column;

        i {
            font-size: 5rem;
        }

        h6 {
            margin: 0.5rem 0;
        }
    }
</style>
