<template>
    <div>
        <RouterButton :buttons="buttons" />
        <LayoutContent>
            <router-view></router-view>
        </LayoutContent>
    </div>
</template>

<script lang="ts" setup>
const buttons = [
    {
        label: 'PHP',
        path: '/websites/runtimes/php',
    },
    {
        label: 'Node.js',
        path: '/websites/runtimes/node',
    },
    {
        label: 'Go',
        path: '/websites/runtimes/go',
    },
    {
        label: 'Java',
        path: '/websites/runtimes/java',
    },
    {
        label: 'Python',
        path: '/websites/runtimes/python',
    },
];
</script>
