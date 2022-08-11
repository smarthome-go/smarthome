<script lang="ts">
    import Button from "@smui/button";
    import DataTable, {
        Head,
        Body,
        Row,
        Cell,
        Pagination,
        Label,
    } from "@smui/data-table";
    import Dialog, { Actions, Content, Header, Title } from "@smui/dialog";
    import IconButton from "@smui/icon-button/src/IconButton.svelte";
    import Select, { Option } from "@smui/select";
    import { onMount } from "svelte";
    import Progress from "../../../components/Progress.svelte";
    import { createSnackbar } from "../../../global";
    import { levels, LogEvent, logs } from "../main";

    // If the dialog should be open or closed
    let open = true;

    // Specifies whether the loading indicator in the logs list should be active or not
    let loading = false;

    let rowsPerPage = 10;
    let minLevel = "TRACE";

    // Pagination
    let start = 0;
    let end = 0;
    let slice: LogEvent[] = [];
    let lastPage = 0;
    let currentPage = 0;

    $: start = currentPage * rowsPerPage;
    $: end = Math.min(start + rowsPerPage, $logs.length);
    $: slice = $logs.slice(start, end);
    $: lastPage = Math.max(Math.ceil($logs.length / rowsPerPage) - 1, 0);

    $: if (currentPage > lastPage) {
        currentPage = lastPage;
    }

    async function deleteRecord(id: number) {
        loading = true;
        try {
            const res = await (
                await fetch(`/api/logs/delete/${id}`, {
                    method: "DELETE",
                })
            ).json();
            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            $logs = res;
        } catch (err) {
            $createSnackbar(`Failed to delete log record: ${err}`);
        }
        loading = false;
    }

    async function fetchLogs() {
        loading = true;
        try {
            const res = await (await fetch("/api/logs/list/all")).json();
            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            $logs = res;
        } catch (err) {
            $createSnackbar(`Failed to load system event logs: ${err}`);
        }
        loading = false;
    }

    // As soon as the component is mounted, fetch the logs
    onMount(fetchLogs);
</script>

<Dialog
    bind:open
    fullscreen
    aria-labelledby="logs-title"
    aria-describedby="logs-content"
>
    <Header>
        <Title id="logs-title">Event Logs</Title>
        <IconButton action="close" class="material-icons">close</IconButton>
    </Header>
    <Content id="logs-content">
        <Progress type="linear" bind:loading />
        <div class="logs">
            <div class="logs__header">
                <IconButton class="material-icons">delete</IconButton>
                <IconButton class="material-icons">delete_forever</IconButton>

                <Select bind:value={minLevel} label="Minimul Level">
                    {#each levels as level}
                        <Option value={level}>
                            <span style:color={level.color}>
                                {level.label}
                            </span>
                        </Option>
                    {/each}
                </Select>
            </div>
            <div class="logs__list">
                <DataTable table$aria-label="Event Log Records" style="width: 100%; height: 40rem;">
                    <Head>
                        <Row>
                            <Cell>Level</Cell>
                            <Cell>Name</Cell>
                            <Cell style="width: 100%;">Description</Cell>
                            <Cell />
                        </Row>
                    </Head>
                    <Body>
                        {#each slice as logEvent (logEvent.id)}
                            <Row>
                                <Cell>{levels[logEvent.level].label}</Cell>
                                <Cell>{logEvent.name}</Cell>
                                <Cell>{logEvent.description}</Cell>
                                <Cell>
                                    <IconButton
                                        class="material-icons"
                                        on:click={() =>
                                            deleteRecord(logEvent.id)}
                                        title="Delete Record">close</IconButton
                                    >
                                </Cell>
                            </Row>
                        {/each}
                    </Body>

                    <Pagination slot="paginate">
                        <svelte:fragment slot="rowsPerPage">
                            <Label>Rows Per Page</Label>
                            <Select
                                variant="outlined"
                                bind:value={rowsPerPage}
                                noLabel
                            >
                                <Option value={10}>10</Option>
                                <Option value={25}>50</Option>
                                <Option value={100}>100</Option>
                            </Select>
                        </svelte:fragment>
                        <svelte:fragment slot="total">
                            {start + 1}-{end} of {$logs.length}
                        </svelte:fragment>

                        <IconButton
                            class="material-icons"
                            action="first-page"
                            title="First page"
                            on:click={() => (currentPage = 0)}
                            disabled={currentPage === 0}>first_page</IconButton
                        >
                        <IconButton
                            class="material-icons"
                            action="prev-page"
                            title="Prev page"
                            on:click={() => currentPage--}
                            disabled={currentPage === 0}
                            >chevron_left</IconButton
                        >
                        <IconButton
                            class="material-icons"
                            action="next-page"
                            title="Next page"
                            on:click={() => currentPage++}
                            disabled={currentPage === lastPage}
                            >chevron_right</IconButton
                        >
                        <IconButton
                            class="material-icons"
                            action="last-page"
                            title="Last page"
                            on:click={() => (currentPage = lastPage)}
                            disabled={currentPage === lastPage}
                            >last_page</IconButton
                        >
                    </Pagination>
                </DataTable>
            </div>
        </div>
    </Content>
    <Actions>
        <Button defaultAction>
            <Label>Close</Label>
        </Button>
    </Actions>
</Dialog>

<style lang="scss">
    .logs {
        &__header {
            display: flex;
        }
    }
</style>
