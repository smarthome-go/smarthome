<script lang="ts">
    import type { homescriptArg } from "../../homescript";

    // Is bound externally
    export let data: homescriptArg;
</script>

<div class="argument">
    {data.data.argKey}
</div>

<style lang="scss">
    .argument {
        width: 3rem;
        height: 1rem;
        border-radius: 0.3rem;
    }
</style>
