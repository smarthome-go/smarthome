<script lang="ts">
    import Button, { Icon } from "@smui/button";
    import DataTable, {
        Head,
        Body,
        Row,
        Cell,
        Pagination,
        Label,
    } from "@smui/data-table";
    import Dialog, {
        Actions,
        Content,
        Header,
        InitialFocus,
        Title,
    } from "@smui/dialog";
    import IconButton from "@smui/icon-button";
    import Select, { Option } from "@smui/select";
    import { onMount } from "svelte";
    import Progress from "../../../components/Progress.svelte";
    import { createSnackbar } from "../../../global";
    import { levels, logEvent, logs } from "../main";

    // If the dialog should be open or closed
    export let open = false;

    // Whether the delete-all log records confirmation dialog should be open or closed
    let flushAllOpen = false;

    // Specifies whether the loading indicator in the logs list should be active or not
    let loading = false;

    let rowsPerPage = 10;
    let minLevel = "INFO";

    // Pagination
    let start = 0;
    let end = 0;
    let slice: logEvent[] = [];
    let lastPage = 0;
    let currentPage = 0;

    $: start = currentPage * rowsPerPage;
    $: end = Math.min(start + rowsPerPage, $logs.length);
    $: slice = $logs
        .slice(start, end)
        .filter(
            (e) => e.level >= levels.findIndex((l) => l.label === minLevel)
        );
    $: lastPage = Math.max(Math.ceil($logs.length / rowsPerPage) - 1, 0);

    $: if (currentPage > lastPage) {
        currentPage = lastPage;
    }

    async function deleteRecord(id: number) {
        loading = true;
        try {
            const res = await (
                await fetch(`/api/logs/delete/id/${id}`, {
                    method: "DELETE",
                })
            ).json();
            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            // Filter out the deleted entry
            $logs = $logs.filter((e) => e.id !== id);
        } catch (err) {
            $createSnackbar(`Failed to delete log record: ${err}`);
        }
        loading = false;
    }

    async function flushAllLogs() {
        loading = true;
        try {
            const res = await (
                await fetch("/api/logs/delete/all", {
                    method: "DELETE",
                })
            ).json();
            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            // Make logs empty in frontend
            $logs = [];
        } catch (err) {
            $createSnackbar(`Failed to flush all log records: ${err}`);
        }
        loading = false;
    }

    async function flushOldLogs() {
        loading = true;
        try {
            const res = await (
                await fetch("/api/logs/delete/old", {
                    method: "DELETE",
                })
            ).json();
            if (res.success !== undefined && !res.success)
                throw Error(res.error);
            loading = false;
            // Must reload logs for changes to take effect
            await fetchLogs();
        } catch (err) {
            loading = false;
            $createSnackbar(`Failed to flush all log records: ${err}`);
        }
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
    aria-labelledby="fullscreen-title"
    aria-describedby="fullscreen-content"
>
    <Header>
        <Title id="fullscreen-title">Event Logs</Title>
        <IconButton action="close" class="material-icons">close</IconButton>
    </Header>
    <Content id="fullscreen-content">
        <Progress type="linear" bind:loading />
        <div class="header">
            <div class="header__left">
                <Select bind:value={minLevel} label="Minimul Log Level">
                    {#each levels as level (level.label)}
                        <Option value={level.label}>
                            <span style:color={level.color}>
                                {level.label}
                            </span>
                        </Option>
                    {/each}
                </Select>
            </div>
            <div class="header__right">
                <Button on:click={flushOldLogs} variant="raised">
                    <Label>Old</Label>
                    <Icon class="material-icons">delete</Icon>
                </Button>
                <Button
                    on:click={() => (flushAllOpen = true)}
                    variant="outlined"
                >
                    <Label>All</Label>
                    <Icon class="material-icons">delete_forever</Icon>
                </Button>
                <IconButton
                    on:click={fetchLogs}
                    title="Refresh"
                    class="material-icons">refresh</IconButton
                >
            </div>
        </div>
        <div class="table">
            <DataTable
                table$aria-label="Event Log Records"
                style="width: 100%; height: 32rem;"
                class="table__component"
            >
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
                            <Cell
                                ><span
                                    class="table__component__row__level"
                                    style:color={levels[logEvent.level].color}
                                >
                                    {levels[logEvent.level].label}
                                </span></Cell
                            >
                            <Cell>{logEvent.name}</Cell>
                            <Cell>{logEvent.description}</Cell>
                            <Cell>
                                <IconButton
                                    size="button"
                                    class="material-icons"
                                    on:click={() => deleteRecord(logEvent.id)}
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
                        disabled={currentPage === 0}>chevron_left</IconButton
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
    </Content>
    <Actions>
        <Button defaultAction>Close</Button>
    </Actions>
    <Dialog
        bind:open={flushAllOpen}
        slot="over"
        aria-labelledby="confirm-title"
        aria-describedby="confirm-description"
    >
        <Header>
            <Title id="confirm-title">Confirmation</Title>
        </Header>
        <Content id="confirm-description"
                 >You are about to delete all logs. Do you want to proceed?
        </Content
        >
        <Actions>
            <Button>
                <Label on:click={flushAllLogs}>Delete</Label>
            </Button>
            <Button defaultAction use={[InitialFocus]}>
                <Label>Cancel</Label>
            </Button>
        </Actions>
    </Dialog>
</Dialog>

<style lang="scss">
    .header {
        padding-bottom: 1rem;
        display: flex;
        justify-content: space-between;

        &__right {
            display: flex;
            align-items: center;
            gap: 1rem;
        }
    }
    .table {
        :global &__component {
            background-color: var(--clr-height-0-2);

            &__row {
                &__level {
                    font-family: "Jetbrains Mono", monospace;
                    font-size: 0.75rem;
                }
            }
        }
    }
</style>
