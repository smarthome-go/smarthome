<script lang="ts">
    export let id: number
    export let name: string
    export let description: string
    export let priority: number
    export let dueDate: string
    export let createdDate: string
    export let userWasNotified: boolean
</script>

<div class="mdc-elevation--z3">
    <h5>{name}</h5>
    <p>{description}</p>
    <p>{dueDate}</p>
    <p class="text-hint">{createdDate}</p>
</div>

<style lang="scss">
    div {
        background-color: var(--clr-height-1-3);
        display: flex;
        border-radius: .3rem;
    }
</style>
