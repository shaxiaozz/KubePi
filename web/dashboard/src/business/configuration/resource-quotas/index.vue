<template>
  <layout-content header="Resource Quotas">
    <div style="float: left">
      <el-button type="primary" size="small" @click="yamlCreate"
                  v-has-permissions="{scope:'namespace',apiGroup:'',resource:'resourcequotas',verb:'create'}">
        YAML
      </el-button>
      <el-button type="primary" size="small" @click="onCreate" v-has-permissions="{scope:'namespace',apiGroup:'',resource:'resourcequotas',verb:'create'}">
        {{ $t("commons.button.create") }}
      </el-button>
      <el-button type="primary" size="small" :disabled="selects.length===0" @click="onDelete()"
                  v-has-permissions="{scope:'namespace',apiGroup:'',resource:'resourcequotas',verb:'delete'}">
        {{ $t("commons.button.delete") }}
      </el-button>
      <el-button type="primary" size="small"
                   @click="exportToXlsx()" icon="el-icon-download">
          {{ $t("commons.button.export") }}
        </el-button>
    </div>
    <complex-table :data="data" :selects.sync="selects" v-loading="loading" @search="search"
                   :pagination-config="paginationConfig" :search-config="searchConfig">
      <el-table-column type="selection" fix></el-table-column>
      <el-table-column :label="$t('commons.table.name')" prop="metadata.name" show-overflow-tooltip>
        <template v-slot:default="{row}">
          <span class="span-link" @click="openDetail(row)">{{ row.metadata.name }}</span>
        </template>
      </el-table-column>
      <el-table-column :label="$t('business.namespace.namespace')" prop="metadata.namespace">
        <template v-slot:default="{row}">
          {{ row.metadata.namespace }}
        </template>
      </el-table-column>
      <el-table-column label="Request" min-width="100px">
        <template v-slot:default="{row}">
          <el-tag type="info" size="mini" v-if="row.status.used && row.status.hard"> pods: {{ row.status.used.pods }} /
            {{ row.status.hard.pods }}
          </el-tag>
          <br>
          <el-tag type="info" size="mini" v-if="row.status.used && row.status.hard"> cpu:
            <span v-if="row.status.used['requests.cpu']">
              {{ row.status.used["requests.cpu"] }} / {{ row.status.hard["requests.cpu"] }}
            </span>
            <span v-else>
              {{ row.status.used["cpu"] }} / {{ row.status.hard["cpu"] }}
            </span>
          </el-tag>
          <br>
          <el-tag type="info" size="mini" v-if="row.status.used && row.status.hard"> memory:
            <span v-if="row.status.used['requests.memory']">
               {{ row.status.used["requests.memory"] }} / {{ row.status.hard["requests.memory"] }}
            </span>
            <span v-else>
              {{ row.status.used["memory"] }} / {{ row.status.hard["memory"] }}
            </span>
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="Limit" min-width="100px">
        <template v-slot:default="{row}">
          <el-tag type="info" size="mini" v-if="row.status.used && row.status.hard"> cpu:
            {{ row.status.used["limits.cpu"] }} / {{ row.status.hard["limits.cpu"] }}
          </el-tag>
          <br>
          <el-tag type="info" size="mini" v-if="row.status.used && row.status.hard"> memory:
            {{ row.status.used["limits.memory"] }} / {{ row.status.hard["limits.memory"] }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="$t('commons.table.created_time')" prop="metadata.creationTimestamp" fix>
        <template v-slot:default="{row}">
          {{ row.metadata.creationTimestamp | age }}
        </template>
      </el-table-column>
      <ko-table-operations :buttons="buttons" :label="$t('commons.table.action')"></ko-table-operations>
    </complex-table>
  </layout-content>
</template>

<script>
import LayoutContent from "@/components/layout/LayoutContent"
import ComplexTable from "@/components/complex-table"
import {deleteResourceQuota, getResourceQuota, listResourceQuotas} from "@/api/resourcequota"
import {downloadYaml} from "@/utils/actions"
import KoTableOperations from "@/components/ko-table-operations"
import {checkPermissions} from "@/utils/permission"

import writeXlsxFile from "write-excel-file";
import { cpuUnitConvert, memoryUnitConvert, numberConvert } from "@/utils/unitConvert"

export default {
  name: "ResourceQuotas",
  components: { ComplexTable, LayoutContent, KoTableOperations },
  data () {
    return {
      data: [],
      selects: [],
      loading: false,
      cluster: "",
      buttons: [
        {
          label: this.$t("commons.button.edit"),
          icon: "el-icon-edit",
          click: (row) => {
            this.$router.push({
              name: "ResourceQuotaEdit",
              params: { namespace: row.metadata.namespace, name: row.metadata.name },
              query: { yamlShow: false }
            })
          },
          disabled: () => {
            return !checkPermissions({ scope: "namespace", apiGroup: "", resource: "resourcequotas", verb: "update" })
          }
        },
        {
          label: this.$t("commons.button.edit_yaml"),
          icon: "el-icon-edit",
          click: (row) => {
            this.$router.push({
              name: "ResourceQuotaEdit",
              params: { namespace: row.metadata.namespace, name: row.metadata.name },
              query: { yamlShow: true }
            })
          },
          disabled: () => {
            return !checkPermissions({ scope: "namespace", apiGroup: "", resource: "resourcequotas", verb: "update" })
          }
        },
        {
          label: this.$t("commons.button.download_yaml"),
          icon: "el-icon-download",
          click: (row) => {
            downloadYaml(row.metadata.name + ".yml", getResourceQuota(this.cluster, row.metadata.namespace, row.metadata.name))
          }
        },
        {
          label: this.$t("commons.button.delete"),
          icon: "el-icon-delete",
          click: (row) => {
            this.onDelete(row)
          },
          disabled: () => {
            return !checkPermissions({ scope: "namespace", apiGroup: "", resource: "resourcequotas", verb: "delete" })
          }
        },
      ],
      paginationConfig: {
        currentPage: 1,
        pageSize: 10,
        total: 0,
      },
      searchConfig: {
        keywords: ""
      }
    }
  },
  methods: {
    search (resetPage) {
      this.loading = true
      if (resetPage) {
        this.paginationConfig.currentPage = 1
      }
      listResourceQuotas(this.cluster, true, this.searchConfig.keywords, this.paginationConfig.currentPage, this.paginationConfig.pageSize).then(res => {
        this.data = res.items
        this.paginationConfig.total = res.total
        this.loading = false
      })
    },
    onCreate () {
      this.$router.push({ name: "ResourceQuotaCreate", query: { yamlShow: false } })
    },
    yamlCreate () {
      this.$router.push({
        name: "ResourceQuotaCreateYaml",
        query: { type: "resourcequotas" },
      })
    },
    onDelete (row) {
      this.$confirm(
        this.$t("commons.confirm_message.delete"),
        this.$t("commons.message_box.prompt"), {
          confirmButtonText: this.$t("commons.button.confirm"),
          cancelButtonText: this.$t("commons.button.cancel"),
          type: "warning",
        }).then(() => {
        this.ps = []
        if (row) {
          this.ps.push(deleteResourceQuota(this.cluster, row.metadata.namespace, row.metadata.name))
        } else {
          if (this.selects.length > 0) {
            for (const select of this.selects) {
              this.ps.push(deleteResourceQuota(this.cluster, select.metadata.namespace, select.metadata.name))
            }
          }
        }
        if (this.ps.length !== 0) {
          Promise.all(this.ps)
            .then(() => {
              this.search(true)
              this.$message({
                type: "success",
                message: this.$t("commons.msg.delete_success"),
              })
            })
            .catch(() => {
              this.search(true)
            })
        }
      })
    },
    openDetail (row) {
      this.$router.push({
        name: "ResourceQuotaDetail",
        params: { namespace: row.metadata.namespace, name: row.metadata.name }
      })
    },
    async exportToXlsx(){
      const schema = [
       {
        column: this.$t("commons.table.name"),
        type: String,
        value: (row) => row.metadata.name,
       },
       {
        column: this.$t("business.namespace.namespace"),
        type: String,
        value: (row) => row.metadata.namespace,
       },
       {
        column: "Requests Pods Used",
        type: Number,
        value: (row) => {
          return row.status.used.pods ? Number(row.status.used.pods) : 0
        },
       },
       {
        column: "Requests Pods hard",
        type: Number,
        value: (row) => {
          return row.status.hard.pods ? Number(row.status.hard.pods) : 0
        },
       },
       {
        column: "Requests Cpu Used(m)",
        type: Number,
        value: (row) => {
          if(row.status.used["requests.cpu"])
          return cpuUnitConvert(row.status.used["requests.cpu"])
          if(row.status.used["cpu"])
          return cpuUnitConvert(row.status.used["cpu"])
          return 0
        },
       },
       {
        column: "Requests Cpu hard(m)",
        type: Number,
        value: (row) => {
          if(row.status.hard["requests.cpu"])
          return cpuUnitConvert(row.status.hard["requests.cpu"])
          if(row.status.hard["cpu"])
          return cpuUnitConvert(row.status.hard["cpu"])
          return 0
        },
       },
       {
        column: "Requests Mem Used(Mi)",
        type: Number,
        value: (row) => {
          if(row.status.used["requests.memory"])
          return memoryUnitConvert(row.status.used["requests.memory"])
          if(row.status.used["memory"])
          return memoryUnitConvert(row.status.used["memory"])
          return 0
        },
       },
       {
        column: "Requests Mem hard(Mi)",
        type: Number,
        value: (row) => {
          if(row.status.hard["requests.memory"])
          return memoryUnitConvert(row.status.hard["requests.memory"])
          if(row.status.hard["memory"])
          return memoryUnitConvert(row.status.hard["memory"])
          return 0
        },
       },
       {
        column: "Limits Cpu Used(m)",
        type: Number,
        value: (row) => {
          if (row.status.used["limits.cpu"])
          return cpuUnitConvert(row.status.used["limits.cpu"])
          return 0
        },
       },
       {
        column: "Limits Cpu hard(m)",
        type: Number,
        value: (row) => {
          if (row.status.hard["limits.cpu"])
          return cpuUnitConvert(row.status.hard["limits.cpu"])
          return 0
        },
       },
       {
        column: "Limits Mem Used(Mi)",
        type: Number,
        value: (row) => {
          if (row.status.used["limits.memory"])
          return memoryUnitConvert(row.status.used["limits.memory"])
          return 0
        },
       },
       {
        column: "Limits Mem hard(Mi)",
        type: Number,
        value: (row) => {
          if (row.status.hard["limits.memory"])
          return memoryUnitConvert(row.status.hard["limits.memory"])
          return 0
        },
       },
       {
        column: "configmaps Used",
        type: Number,
        value: (row) => {
          return numberConvert(row.status.used.configmaps||'')
        },
       },
       {
        column: "configmaps hard",
        type: Number,
        value: (row) => {
          return numberConvert(row.status.hard.configmaps||'')
        },
       },
       {
        column: "secrets Used",
        type: Number,
        value: (row) => {
          return numberConvert(row.status.used.secrets||'')
        },
       },
       {
        column: "secrets hard",
        type: Number,
        value: (row) => {
          return numberConvert(row.status.hard.secrets||'')
        },
       },
       {
        column: "services Used",
        type: Number,
        value: (row) => {
          return numberConvert(row.status.used.services||'')
        },
       },
       {
        column: "services hard",
        type: Number,
        value: (row) => {
          return numberConvert(row.status.hard.services||'')
        },
       },
      ];
      const data=await listResourceQuotas(this.cluster, true, this.searchConfig.keywords)
      await writeXlsxFile((data.items||[]), {
         schema,
         fileName: "resource-quotas.xlsx",
      });
    }
  },
  created () {
    this.cluster = this.$route.query.cluster
    this.search()
  }
}
</script>

<style scoped>

</style>
