<script lang="ts">
    import Button, { Label } from '@smui/button'
    import Dialog, { Actions, Content, InitialFocus, Title } from '@smui/dialog'
    import Progress from '../../../../../src/components/Progress.svelte'
    import { fetchHardwareNodes, loading, hardwareNodesLoaded, hardwareNodes } from './main'

    let open = false

    export let id: string
    export let name: string
    export let watts: number
    export let targetNode: string
    let targetNodeLabel = undefined

    export function show() {
        open = true

        if (!$hardwareNodesLoaded) {
            fetchHardwareNodes().then(() => {
                setNodeLabel()
            })
        } else {
            setNodeLabel()
        }
    }

    function setNodeLabel() {
        if (targetNode === null || targetNode === undefined) {
            targetNodeLabel = 'None'
        } else {
            targetNodeLabel = $hardwareNodes.find(h => {
                if (h === undefined || h === null) {
                    return false
                }
                return h.url === targetNode
            }).name
        }
    }
</script>

<Dialog bind:open aria-labelledby="title" aria-describedby="content">
    <Title id="title">Switch Information</Title>
    <Content id="content">
        ID: <code>{id}</code>
        <br />
        Name: {name}
        <br />
        Watts: {watts}
        <br />
        {#if $hardwareNodesLoaded}
            Target Node: {targetNodeLabel}
        {:else}
            <Progress bind:loading={$loading} />
        {/if}
        <br />
    </Content>
    <Actions>
        <Button defaultAction use={[InitialFocus]}>
            <Label>Done</Label>
        </Button>
    </Actions>
</Dialog>

<style style="scss">
    code {
        background-color: var(--clr-height-0-3);
        padding: 0.1rem 0.5rem;
        border-radius: 0.3rem;
    }
</style>
