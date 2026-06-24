#!/usr/bin/env python3
"""
多省天气数据采集脚本（Open-Meteo 免费API）— 直写 PostgreSQL
覆盖：福建、江西、安徽、江苏、上海
采集：逐小时气象 + 逐日水文
存储：PostgreSQL ptis.md_weather_hourly + md_weather_hydrology_daily

用法：
  python fetch_weather_data.py           # 采集全部省份
  python fetch_weather_data.py fujian    # 仅福建
  python fetch_weather_data.py jiangxi   # 仅江西
  python fetch_weather_data.py anhui     # 仅安徽
  python fetch_weather_data.py jiangsu   # 仅江苏
  python fetch_weather_data.py shanghai  # 仅上海

环境变量（可选覆盖默认值）:
  PG_HOST, PG_PORT, PG_USER, PG_PASSWORD, PG_DB
"""
import os, sys, json, time
from datetime import datetime, date, timedelta
import psycopg2
from psycopg2.extras import execute_values
import urllib.request

PG = {
    "host": os.environ.get("PG_HOST", "localhost"),
    "port": int(os.environ.get("PG_PORT", "5432")),
    "user": os.environ.get("PG_USER", "ptis"),
    "password": os.environ.get("PG_PASSWORD", os.environ.get("POSTGRES_PASSWORD", "ptis")),
    "dbname": os.environ.get("PG_DB", "ptis"),
}

# ── 省份 → 区县坐标 ────────────────────────────────────

PROVINCES = {
    "fujian": [
        {'code': 'FZ_GL', 'name': '鼓楼区', 'city': '福州市', 'lat': 26.08, 'lon': 119.30},
        {'code': 'FZ_TJ', 'name': '台江区', 'city': '福州市', 'lat': 26.06, 'lon': 119.31},
        {'code': 'FZ_CS', 'name': '仓山区', 'city': '福州市', 'lat': 26.05, 'lon': 119.32},
        {'code': 'FZ_MW', 'name': '马尾区', 'city': '福州市', 'lat': 26.00, 'lon': 119.45},
        {'code': 'FZ_JA', 'name': '晋安区', 'city': '福州市', 'lat': 26.08, 'lon': 119.33},
        {'code': 'FZ_CL', 'name': '长乐区', 'city': '福州市', 'lat': 25.96, 'lon': 119.52},
        {'code': 'FZ_FQ', 'name': '福清市', 'city': '福州市', 'lat': 25.73, 'lon': 119.38},
        {'code': 'FZ_MH', 'name': '闽侯县', 'city': '福州市', 'lat': 26.15, 'lon': 118.97},
        {'code': 'FZ_LJ', 'name': '连江县', 'city': '福州市', 'lat': 26.20, 'lon': 119.53},
        {'code': 'FZ_LY', 'name': '罗源县', 'city': '福州市', 'lat': 26.49, 'lon': 119.55},
        {'code': 'FZ_MQ', 'name': '闽清县', 'city': '福州市', 'lat': 26.22, 'lon': 118.86},
        {'code': 'FZ_YT', 'name': '永泰县', 'city': '福州市', 'lat': 25.87, 'lon': 118.94},
        {'code': 'FZ_PT', 'name': '平潭县', 'city': '福州市', 'lat': 25.50, 'lon': 119.79},
        {'code': 'XM_SM', 'name': '思明区', 'city': '厦门市', 'lat': 24.45, 'lon': 118.08},
        {'code': 'XM_HC', 'name': '海沧区', 'city': '厦门市', 'lat': 24.48, 'lon': 118.03},
        {'code': 'XM_HL', 'name': '湖里区', 'city': '厦门市', 'lat': 24.51, 'lon': 118.15},
        {'code': 'XM_JM', 'name': '集美区', 'city': '厦门市', 'lat': 24.57, 'lon': 118.10},
        {'code': 'XM_TA', 'name': '同安区', 'city': '厦门市', 'lat': 24.72, 'lon': 118.15},
        {'code': 'XM_XA', 'name': '翔安区', 'city': '厦门市', 'lat': 24.62, 'lon': 118.25},
        {'code': 'PT_CX', 'name': '城厢区', 'city': '莆田市', 'lat': 25.42, 'lon': 118.93},
        {'code': 'PT_HJ', 'name': '涵江区', 'city': '莆田市', 'lat': 25.46, 'lon': 119.07},
        {'code': 'PT_LC', 'name': '荔城区', 'city': '莆田市', 'lat': 25.43, 'lon': 119.02},
        {'code': 'PT_XY', 'name': '秀屿区', 'city': '莆田市', 'lat': 25.32, 'lon': 119.10},
        {'code': 'PT_XYX', 'name': '仙游县', 'city': '莆田市', 'lat': 25.36, 'lon': 118.69},
        {'code': 'SM_SY', 'name': '三元区', 'city': '三明市', 'lat': 26.23, 'lon': 117.58},
        {'code': 'SM_SX', 'name': '沙县区', 'city': '三明市', 'lat': 26.40, 'lon': 117.79},
        {'code': 'SM_YA', 'name': '永安市', 'city': '三明市', 'lat': 25.97, 'lon': 117.36},
        {'code': 'SM_MX', 'name': '明溪县', 'city': '三明市', 'lat': 26.36, 'lon': 117.20},
        {'code': 'SM_QL', 'name': '清流县', 'city': '三明市', 'lat': 26.18, 'lon': 116.82},
        {'code': 'SM_NH', 'name': '宁化县', 'city': '三明市', 'lat': 26.26, 'lon': 116.64},
        {'code': 'SM_DT', 'name': '大田县', 'city': '三明市', 'lat': 25.69, 'lon': 117.85},
        {'code': 'SM_YX', 'name': '尤溪县', 'city': '三明市', 'lat': 26.17, 'lon': 118.19},
        {'code': 'SM_JL', 'name': '将乐县', 'city': '三明市', 'lat': 26.73, 'lon': 117.47},
        {'code': 'SM_TN', 'name': '泰宁县', 'city': '三明市', 'lat': 26.90, 'lon': 117.18},
        {'code': 'SM_JN', 'name': '建宁县', 'city': '三明市', 'lat': 26.84, 'lon': 116.68},
        {'code': 'QZ_LC', 'name': '鲤城区', 'city': '泉州市', 'lat': 24.91, 'lon': 118.59},
        {'code': 'QZ_FZ', 'name': '丰泽区', 'city': '泉州市', 'lat': 24.91, 'lon': 118.61},
        {'code': 'QZ_LJ', 'name': '洛江区', 'city': '泉州市', 'lat': 24.94, 'lon': 118.67},
        {'code': 'QZ_QG', 'name': '泉港区', 'city': '泉州市', 'lat': 25.13, 'lon': 118.72},
        {'code': 'QZ_JJ', 'name': '晋江市', 'city': '泉州市', 'lat': 24.78, 'lon': 118.55},
        {'code': 'QZ_SS', 'name': '石狮市', 'city': '泉州市', 'lat': 24.73, 'lon': 118.65},
        {'code': 'QZ_NA', 'name': '南安市', 'city': '泉州市', 'lat': 24.96, 'lon': 118.39},
        {'code': 'QZ_HA', 'name': '惠安县', 'city': '泉州市', 'lat': 25.04, 'lon': 118.80},
        {'code': 'QZ_AX', 'name': '安溪县', 'city': '泉州市', 'lat': 25.06, 'lon': 118.18},
        {'code': 'QZ_YC', 'name': '永春县', 'city': '泉州市', 'lat': 25.32, 'lon': 118.29},
        {'code': 'QZ_DH', 'name': '德化县', 'city': '泉州市', 'lat': 25.49, 'lon': 118.24},
        {'code': 'ZZ_XC', 'name': '芗城区', 'city': '漳州市', 'lat': 24.51, 'lon': 117.65},
        {'code': 'ZZ_LW', 'name': '龙文区', 'city': '漳州市', 'lat': 24.50, 'lon': 117.71},
        {'code': 'ZZ_LH', 'name': '龙海区', 'city': '漳州市', 'lat': 24.45, 'lon': 117.82},
        {'code': 'ZZ_YX', 'name': '云霄县', 'city': '漳州市', 'lat': 23.96, 'lon': 117.34},
        {'code': 'ZZ_ZP', 'name': '漳浦县', 'city': '漳州市', 'lat': 24.05, 'lon': 117.61},
        {'code': 'ZZ_ZA', 'name': '诏安县', 'city': '漳州市', 'lat': 23.71, 'lon': 117.18},
        {'code': 'ZZ_CT', 'name': '长泰县', 'city': '漳州市', 'lat': 24.63, 'lon': 117.76},
        {'code': 'ZZ_NJ', 'name': '南靖县', 'city': '漳州市', 'lat': 24.51, 'lon': 117.36},
        {'code': 'ZZ_PK', 'name': '平和县', 'city': '漳州市', 'lat': 24.36, 'lon': 117.31},
        {'code': 'ZZ_HA', 'name': '华安县', 'city': '漳州市', 'lat': 25.00, 'lon': 117.53},
        {'code': 'NP_YB', 'name': '延平区', 'city': '南平市', 'lat': 26.64, 'lon': 118.18},
        {'code': 'NP_JY', 'name': '建阳区', 'city': '南平市', 'lat': 27.34, 'lon': 118.12},
        {'code': 'NP_SC', 'name': '邵武市', 'city': '南平市', 'lat': 27.34, 'lon': 117.49},
        {'code': 'NP_WY', 'name': '武夷山市', 'city': '南平市', 'lat': 27.76, 'lon': 118.03},
        {'code': 'NP_JS', 'name': '建瓯市', 'city': '南平市', 'lat': 27.05, 'lon': 118.32},
        {'code': 'NP_SY', 'name': '顺昌县', 'city': '南平市', 'lat': 26.80, 'lon': 117.81},
        {'code': 'NP_PJ', 'name': '浦城县', 'city': '南平市', 'lat': 27.92, 'lon': 118.54},
        {'code': 'NP_GC', 'name': '光泽县', 'city': '南平市', 'lat': 27.54, 'lon': 117.34},
        {'code': 'NP_SB', 'name': '松溪县', 'city': '南平市', 'lat': 27.53, 'lon': 118.78},
        {'code': 'NP_ZH', 'name': '政和县', 'city': '南平市', 'lat': 27.38, 'lon': 118.86},
        {'code': 'LY_XH', 'name': '新罗区', 'city': '龙岩市', 'lat': 25.10, 'lon': 117.03},
        {'code': 'LY_YD', 'name': '永定区', 'city': '龙岩市', 'lat': 24.73, 'lon': 116.73},
        {'code': 'LY_ZP', 'name': '漳平市', 'city': '龙岩市', 'lat': 25.30, 'lon': 117.42},
        {'code': 'LY_CT', 'name': '长汀县', 'city': '龙岩市', 'lat': 25.86, 'lon': 116.36},
        {'code': 'LY_SH', 'name': '上杭县', 'city': '龙岩市', 'lat': 25.05, 'lon': 116.42},
        {'code': 'LY_WY', 'name': '武平县', 'city': '龙岩市', 'lat': 25.10, 'lon': 116.10},
        {'code': 'LY_LH', 'name': '连城县', 'city': '龙岩市', 'lat': 25.71, 'lon': 116.75},
        {'code': 'ND_JC', 'name': '蕉城区', 'city': '宁德市', 'lat': 26.66, 'lon': 119.53},
        {'code': 'ND_FA', 'name': '福安市', 'city': '宁德市', 'lat': 27.09, 'lon': 119.65},
        {'code': 'ND_FD', 'name': '福鼎市', 'city': '宁德市', 'lat': 27.33, 'lon': 120.20},
        {'code': 'ND_LS', 'name': '霞浦县', 'city': '宁德市', 'lat': 26.89, 'lon': 120.00},
        {'code': 'ND_GT', 'name': '古田县', 'city': '宁德市', 'lat': 26.58, 'lon': 118.74},
        {'code': 'ND_PT', 'name': '屏南县', 'city': '宁德市', 'lat': 26.91, 'lon': 118.99},
        {'code': 'ND_SC', 'name': '寿宁县', 'city': '宁德市', 'lat': 27.46, 'lon': 119.51},
        {'code': 'ND_ZR', 'name': '周宁县', 'city': '宁德市', 'lat': 27.10, 'lon': 119.34},
        {'code': 'ND_ZN', 'name': '柘荣县', 'city': '宁德市', 'lat': 27.24, 'lon': 119.90},
    ],
    "jiangxi": [
        {'code': 'NC_DH', 'name': '东湖区', 'city': '南昌市', 'lat': 28.68, 'lon': 115.90},
        {'code': 'NC_XH', 'name': '西湖区', 'city': '南昌市', 'lat': 28.66, 'lon': 115.87},
        {'code': 'NC_QS', 'name': '青云谱区', 'city': '南昌市', 'lat': 28.63, 'lon': 115.92},
        {'code': 'NC_QY', 'name': '青山湖区', 'city': '南昌市', 'lat': 28.68, 'lon': 115.96},
        {'code': 'NC_WL', 'name': '湾里区', 'city': '南昌市', 'lat': 28.78, 'lon': 115.76},
        {'code': 'NC_NC', 'name': '南昌县', 'city': '南昌市', 'lat': 28.55, 'lon': 115.94},
        {'code': 'NC_XJ', 'name': '新建区', 'city': '南昌市', 'lat': 28.68, 'lon': 115.81},
        {'code': 'NC_AC', 'name': '安义县', 'city': '南昌市', 'lat': 28.84, 'lon': 115.55},
        {'code': 'NC_JX', 'name': '进贤县', 'city': '南昌市', 'lat': 28.38, 'lon': 116.24},
        {'code': 'JZ_YZ', 'name': '昌江区', 'city': '景德镇市', 'lat': 29.13, 'lon': 117.19},
        {'code': 'JZ_ZS', 'name': '珠山区', 'city': '景德镇市', 'lat': 29.30, 'lon': 117.28},
        {'code': 'JZ_FS', 'name': '浮梁县', 'city': '景德镇市', 'lat': 29.39, 'lon': 117.22},
        {'code': 'JZ_LR', 'name': '乐平市', 'city': '景德镇市', 'lat': 28.96, 'lon': 117.13},
        {'code': 'PX_AT', 'name': '安源区', 'city': '萍乡市', 'lat': 27.62, 'lon': 113.85},
        {'code': 'PX_XD', 'name': '湘东区', 'city': '萍乡市', 'lat': 27.64, 'lon': 113.73},
        {'code': 'PX_LS', 'name': '莲花县', 'city': '萍乡市', 'lat': 27.13, 'lon': 113.96},
        {'code': 'PX_SK', 'name': '上栗县', 'city': '萍乡市', 'lat': 27.88, 'lon': 113.80},
        {'code': 'PX_LH', 'name': '芦溪县', 'city': '萍乡市', 'lat': 27.64, 'lon': 114.03},
        {'code': 'JD_ZS', 'name': '珠山区', 'city': '九江市', 'lat': 29.73, 'lon': 115.99},
        {'code': 'JD_SL', 'name': '濂溪区', 'city': '九江市', 'lat': 29.67, 'lon': 115.99},
        {'code': 'JD_FC', 'name': '柴桑区', 'city': '九江市', 'lat': 29.61, 'lon': 115.91},
        {'code': 'JD_WS', 'name': '武宁县', 'city': '九江市', 'lat': 29.26, 'lon': 115.10},
        {'code': 'JD_XS', 'name': '修水县', 'city': '九江市', 'lat': 29.03, 'lon': 114.57},
        {'code': 'JD_YX', 'name': '永修县', 'city': '九江市', 'lat': 29.02, 'lon': 115.82},
        {'code': 'JD_DG', 'name': '德安县', 'city': '九江市', 'lat': 29.33, 'lon': 115.76},
        {'code': 'JD_DC', 'name': '都昌县', 'city': '九江市', 'lat': 29.27, 'lon': 116.21},
        {'code': 'JD_HK', 'name': '湖口县', 'city': '九江市', 'lat': 29.74, 'lon': 116.22},
        {'code': 'JD_PC', 'name': '彭泽县', 'city': '九江市', 'lat': 29.90, 'lon': 116.55},
        {'code': 'JD_RC', 'name': '瑞昌市', 'city': '九江市', 'lat': 29.68, 'lon': 115.67},
        {'code': 'JD_LS', 'name': '庐山市', 'city': '九江市', 'lat': 29.45, 'lon': 116.05},
        {'code': 'JD_GC', 'name': '共青城市', 'city': '九江市', 'lat': 29.25, 'lon': 115.80},
        {'code': 'XY_SW', 'name': '渝水区', 'city': '新余市', 'lat': 27.80, 'lon': 114.93},
        {'code': 'XY_FD', 'name': '分宜县', 'city': '新余市', 'lat': 27.81, 'lon': 114.69},
        {'code': 'YG_YZ', 'name': '月湖区', 'city': '鹰潭市', 'lat': 28.24, 'lon': 117.07},
        {'code': 'YG_YJ', 'name': '余江区', 'city': '鹰潭市', 'lat': 28.21, 'lon': 116.82},
        {'code': 'YG_GB', 'name': '贵溪市', 'city': '鹰潭市', 'lat': 28.30, 'lon': 117.21},
        {'code': 'GZ_ZQ', 'name': '章贡区', 'city': '赣州市', 'lat': 25.87, 'lon': 114.92},
        {'code': 'GZ_NS', 'name': '南康区', 'city': '赣州市', 'lat': 25.66, 'lon': 114.75},
        {'code': 'GZ_GN', 'name': '赣县区', 'city': '赣州市', 'lat': 25.86, 'lon': 115.00},
        {'code': 'GZ_XF', 'name': '信丰县', 'city': '赣州市', 'lat': 25.39, 'lon': 114.92},
        {'code': 'GZ_DA', 'name': '大余县', 'city': '赣州市', 'lat': 25.40, 'lon': 114.36},
        {'code': 'GZ_SN', 'name': '上犹县', 'city': '赣州市', 'lat': 25.79, 'lon': 114.55},
        {'code': 'GZ_CY', 'name': '崇义县', 'city': '赣州市', 'lat': 25.68, 'lon': 114.31},
        {'code': 'GZ_AN', 'name': '安远县', 'city': '赣州市', 'lat': 25.14, 'lon': 115.41},
        {'code': 'GZ_LN', 'name': '龙南市', 'city': '赣州市', 'lat': 24.91, 'lon': 114.79},
        {'code': 'GZ_DN', 'name': '定南县', 'city': '赣州市', 'lat': 24.78, 'lon': 115.03},
        {'code': 'GZ_QN', 'name': '全南县', 'city': '赣州市', 'lat': 24.74, 'lon': 114.68},
        {'code': 'GZ_NC', 'name': '宁都县', 'city': '赣州市', 'lat': 26.47, 'lon': 116.01},
        {'code': 'GZ_YD', 'name': '于都县', 'city': '赣州市', 'lat': 25.95, 'lon': 115.42},
        {'code': 'GZ_XG', 'name': '兴国县', 'city': '赣州市', 'lat': 26.34, 'lon': 115.36},
        {'code': 'GZ_RJ', 'name': '瑞金市', 'city': '赣州市', 'lat': 25.89, 'lon': 116.03},
        {'code': 'GZ_HW', 'name': '会昌县', 'city': '赣州市', 'lat': 25.60, 'lon': 115.79},
        {'code': 'GZ_XN', 'name': '寻乌县', 'city': '赣州市', 'lat': 24.96, 'lon': 115.65},
        {'code': 'GZ_SC', 'name': '石城县', 'city': '赣州市', 'lat': 26.33, 'lon': 116.35},
        {'code': 'JY_QZ', 'name': '吉州区', 'city': '吉安市', 'lat': 27.10, 'lon': 114.99},
        {'code': 'JY_QQ', 'name': '青原区', 'city': '吉安市', 'lat': 27.08, 'lon': 115.14},
        {'code': 'JY_JA', 'name': '吉安县', 'city': '吉安市', 'lat': 27.04, 'lon': 114.91},
        {'code': 'JY_JS', 'name': '吉水县', 'city': '吉安市', 'lat': 27.21, 'lon': 115.13},
        {'code': 'JY_XJ', 'name': '峡江县', 'city': '吉安市', 'lat': 27.58, 'lon': 115.32},
        {'code': 'JY_XF', 'name': '新干县', 'city': '吉安市', 'lat': 27.74, 'lon': 115.39},
        {'code': 'JY_YF', 'name': '永丰县', 'city': '吉安市', 'lat': 27.33, 'lon': 115.42},
        {'code': 'JY_TY', 'name': '泰和县', 'city': '吉安市', 'lat': 26.79, 'lon': 114.91},
        {'code': 'JY_ST', 'name': '遂川县', 'city': '吉安市', 'lat': 26.32, 'lon': 114.50},
        {'code': 'JY_WA', 'name': '万安县', 'city': '吉安市', 'lat': 26.46, 'lon': 114.79},
        {'code': 'JY_AF', 'name': '安福县', 'city': '吉安市', 'lat': 27.39, 'lon': 114.62},
        {'code': 'JY_YC', 'name': '永新县', 'city': '吉安市', 'lat': 26.96, 'lon': 114.24},
        {'code': 'JY_JN', 'name': '井冈山市', 'city': '吉安市', 'lat': 26.58, 'lon': 114.17},
        {'code': 'FZ_LC', 'name': '临川区', 'city': '抚州市', 'lat': 27.95, 'lon': 116.36},
        {'code': 'FZ_DC', 'name': '东乡区', 'city': '抚州市', 'lat': 28.24, 'lon': 116.60},
        {'code': 'FZ_NC', 'name': '南城县', 'city': '抚州市', 'lat': 27.55, 'lon': 116.64},
        {'code': 'FZ_LC', 'name': '黎川县', 'city': '抚州市', 'lat': 27.30, 'lon': 116.91},
        {'code': 'FZ_NS', 'name': '南丰县', 'city': '抚州市', 'lat': 27.22, 'lon': 116.54},
        {'code': 'FZ_CJ', 'name': '崇仁县', 'city': '抚州市', 'lat': 27.76, 'lon': 116.06},
        {'code': 'FZ_LC', 'name': '乐安县', 'city': '抚州市', 'lat': 27.43, 'lon': 115.84},
        {'code': 'FZ_YH', 'name': '宜黄县', 'city': '抚州市', 'lat': 27.55, 'lon': 116.22},
        {'code': 'FZ_JX', 'name': '金溪县', 'city': '抚州市', 'lat': 27.92, 'lon': 116.76},
        {'code': 'FZ_ZX', 'name': '资溪县', 'city': '抚州市', 'lat': 27.71, 'lon': 117.06},
        {'code': 'FZ_GC', 'name': '广昌县', 'city': '抚州市', 'lat': 26.84, 'lon': 116.32},
        {'code': 'SR_XS', 'name': '信州区', 'city': '上饶市', 'lat': 28.44, 'lon': 117.96},
        {'code': 'SR_GZ', 'name': '广丰区', 'city': '上饶市', 'lat': 28.44, 'lon': 118.19},
        {'code': 'SR_GF', 'name': '广信区', 'city': '上饶市', 'lat': 28.45, 'lon': 117.91},
        {'code': 'SR_YG', 'name': '玉山县', 'city': '上饶市', 'lat': 28.68, 'lon': 118.25},
        {'code': 'SR_LF', 'name': '铅山县', 'city': '上饶市', 'lat': 28.32, 'lon': 117.71},
        {'code': 'SR_HY', 'name': '横峰县', 'city': '上饶市', 'lat': 28.42, 'lon': 117.60},
        {'code': 'SR_YF', 'name': '弋阳县', 'city': '上饶市', 'lat': 28.40, 'lon': 117.45},
        {'code': 'SR_YS', 'name': '余干县', 'city': '上饶市', 'lat': 28.70, 'lon': 116.70},
        {'code': 'SR_PY', 'name': '鄱阳县', 'city': '上饶市', 'lat': 29.00, 'lon': 116.68},
        {'code': 'SR_WN', 'name': '万年县', 'city': '上饶市', 'lat': 28.70, 'lon': 117.07},
        {'code': 'SR_WS', 'name': '婺源县', 'city': '上饶市', 'lat': 29.25, 'lon': 117.86},
        {'code': 'SR_DX', 'name': '德兴市', 'city': '上饶市', 'lat': 28.95, 'lon': 117.58},
    ],
    "anhui": [
        {'code': 'HF_LY', 'name': '庐阳区', 'city': '合肥市', 'lat': 31.87, 'lon': 117.26},
        {'code': 'HF_YH', 'name': '瑶海区', 'city': '合肥市', 'lat': 31.86, 'lon': 117.31},
        {'code': 'HF_SS', 'name': '蜀山区', 'city': '合肥市', 'lat': 31.86, 'lon': 117.21},
        {'code': 'HF_BH', 'name': '包河区', 'city': '合肥市', 'lat': 31.80, 'lon': 117.31},
        {'code': 'HF_CD', 'name': '长丰县', 'city': '合肥市', 'lat': 32.08, 'lon': 117.17},
        {'code': 'HF_FD', 'name': '肥东县', 'city': '合肥市', 'lat': 31.89, 'lon': 117.47},
        {'code': 'HF_FX', 'name': '肥西县', 'city': '合肥市', 'lat': 31.71, 'lon': 117.16},
        {'code': 'HF_LJ', 'name': '庐江县', 'city': '合肥市', 'lat': 31.26, 'lon': 117.29},
        {'code': 'HF_CC', 'name': '巢湖市', 'city': '合肥市', 'lat': 31.60, 'lon': 117.87},
        {'code': 'WH_JZ', 'name': '镜湖区', 'city': '芜湖市', 'lat': 31.35, 'lon': 118.43},
        {'code': 'WH_YJ', 'name': '弋江区', 'city': '芜湖市', 'lat': 31.31, 'lon': 118.37},
        {'code': 'WH_JJ', 'name': '鸠江区', 'city': '芜湖市', 'lat': 31.37, 'lon': 118.39},
        {'code': 'WH_SA', 'name': '三山区', 'city': '芜湖市', 'lat': 31.21, 'lon': 118.23},
        {'code': 'WH_WZ', 'name': '芜湖县', 'city': '芜湖市', 'lat': 31.15, 'lon': 118.57},
        {'code': 'WH_FN', 'name': '繁昌县', 'city': '芜湖市', 'lat': 31.08, 'lon': 118.20},
        {'code': 'WH_NL', 'name': '南陵县', 'city': '芜湖市', 'lat': 30.92, 'lon': 118.33},
        {'code': 'WH_WW', 'name': '无为县', 'city': '芜湖市', 'lat': 31.30, 'lon': 117.90},
        {'code': 'BB_LS', 'name': '龙子湖区', 'city': '蚌埠市', 'lat': 32.94, 'lon': 117.39},
        {'code': 'BB_BH', 'name': '蚌山区', 'city': '蚌埠市', 'lat': 32.91, 'lon': 117.37},
        {'code': 'BB_YJ', 'name': '禹会区', 'city': '蚌埠市', 'lat': 32.94, 'lon': 117.35},
        {'code': 'BB_HZ', 'name': '淮上区', 'city': '蚌埠市', 'lat': 32.96, 'lon': 117.36},
        {'code': 'BB_HY', 'name': '怀远县', 'city': '蚌埠市', 'lat': 32.95, 'lon': 116.99},
        {'code': 'BB_WJ', 'name': '五河县', 'city': '蚌埠市', 'lat': 33.14, 'lon': 117.89},
        {'code': 'BB_GZ', 'name': '固镇县', 'city': '蚌埠市', 'lat': 33.32, 'lon': 117.32},
        {'code': 'HN_DZ', 'name': '大通区', 'city': '淮南市', 'lat': 32.63, 'lon': 117.05},
        {'code': 'HN_TJ', 'name': '田家庵区', 'city': '淮南市', 'lat': 32.65, 'lon': 117.02},
        {'code': 'HN_XJ', 'name': '谢家集区', 'city': '淮南市', 'lat': 32.60, 'lon': 116.86},
        {'code': 'HN_BG', 'name': '八公山区', 'city': '淮南市', 'lat': 32.63, 'lon': 116.83},
        {'code': 'HN_PJ', 'name': '潘集区', 'city': '淮南市', 'lat': 32.77, 'lon': 116.84},
        {'code': 'HN_FZ', 'name': '凤台县', 'city': '淮南市', 'lat': 32.65, 'lon': 116.71},
        {'code': 'HN_SJ', 'name': '寿县', 'city': '淮南市', 'lat': 32.57, 'lon': 116.79},
        {'code': 'MAS_RB', 'name': '雨山区', 'city': '马鞍山市', 'lat': 31.68, 'lon': 118.49},
        {'code': 'MAS_HZ', 'name': '花山区', 'city': '马鞍山市', 'lat': 31.70, 'lon': 118.51},
        {'code': 'MAS_BS', 'name': '博望区', 'city': '马鞍山市', 'lat': 31.55, 'lon': 118.85},
        {'code': 'MAS_DA', 'name': '当涂县', 'city': '马鞍山市', 'lat': 31.55, 'lon': 118.49},
        {'code': 'MAS_HS', 'name': '含山县', 'city': '马鞍山市', 'lat': 31.73, 'lon': 118.11},
        {'code': 'MAS_HW', 'name': '和县', 'city': '马鞍山市', 'lat': 31.69, 'lon': 118.37},
        {'code': 'WH_DS', 'name': '杜集区', 'city': '淮北市', 'lat': 33.98, 'lon': 116.79},
        {'code': 'WH_XQ', 'name': '相山区', 'city': '淮北市', 'lat': 33.97, 'lon': 116.79},
        {'code': 'WH_LJ', 'name': '烈山区', 'city': '淮北市', 'lat': 33.89, 'lon': 116.81},
        {'code': 'WH_SN', 'name': '濉溪县', 'city': '淮北市', 'lat': 33.92, 'lon': 116.77},
        {'code': 'TS_CC', 'name': '铜官区', 'city': '铜陵市', 'lat': 30.93, 'lon': 117.81},
        {'code': 'TS_YA', 'name': '义安区', 'city': '铜陵市', 'lat': 30.95, 'lon': 117.75},
        {'code': 'TS_JQ', 'name': '郊区', 'city': '铜陵市', 'lat': 30.91, 'lon': 117.80},
        {'code': 'TS_ZY', 'name': '枞阳县', 'city': '铜陵市', 'lat': 30.70, 'lon': 117.22},
        {'code': 'AQ_YX', 'name': '迎江区', 'city': '安庆市', 'lat': 30.52, 'lon': 117.06},
        {'code': 'AQ_DZ', 'name': '大观区', 'city': '安庆市', 'lat': 30.55, 'lon': 117.03},
        {'code': 'AQ_YC', 'name': '宜秀区', 'city': '安庆市', 'lat': 30.61, 'lon': 117.09},
        {'code': 'AQ_HN', 'name': '怀宁县', 'city': '安庆市', 'lat': 30.74, 'lon': 116.83},
        {'code': 'AQ_QM', 'name': '潜山市', 'city': '安庆市', 'lat': 30.63, 'lon': 116.58},
        {'code': 'AQ_TM', 'name': '太湖县', 'city': '安庆市', 'lat': 30.45, 'lon': 116.31},
        {'code': 'AQ_SS', 'name': '宿松县', 'city': '安庆市', 'lat': 30.15, 'lon': 116.13},
        {'code': 'AQ_WX', 'name': '望江县', 'city': '安庆市', 'lat': 30.12, 'lon': 116.69},
        {'code': 'AQ_YS', 'name': '岳西县', 'city': '安庆市', 'lat': 30.85, 'lon': 116.36},
        {'code': 'AQ_TT', 'name': '桐城市', 'city': '安庆市', 'lat': 31.04, 'lon': 116.95},
        {'code': 'HS_QS', 'name': '屯溪区', 'city': '黄山市', 'lat': 29.73, 'lon': 118.31},
        {'code': 'HS_HZ', 'name': '黄山区', 'city': '黄山市', 'lat': 30.27, 'lon': 118.13},
        {'code': 'HS_WN', 'name': '徽州区', 'city': '黄山市', 'lat': 29.82, 'lon': 118.34},
        {'code': 'HS_SX', 'name': '歙县', 'city': '黄山市', 'lat': 29.86, 'lon': 118.41},
        {'code': 'HS_XN', 'name': '休宁县', 'city': '黄山市', 'lat': 29.79, 'lon': 118.19},
        {'code': 'HS_YX', 'name': '黟县', 'city': '黄山市', 'lat': 29.92, 'lon': 117.94},
        {'code': 'HS_QM', 'name': '祁门县', 'city': '黄山市', 'lat': 29.85, 'lon': 117.72},
        {'code': 'CZ_LC', 'name': '琅琊区', 'city': '滁州市', 'lat': 32.30, 'lon': 118.31},
        {'code': 'CZ_NQ', 'name': '南谯区', 'city': '滁州市', 'lat': 32.33, 'lon': 118.30},
        {'code': 'CZ_LA', 'name': '来安县', 'city': '滁州市', 'lat': 32.45, 'lon': 118.43},
        {'code': 'CZ_QJ', 'name': '全椒县', 'city': '滁州市', 'lat': 32.10, 'lon': 117.84},
        {'code': 'CZ_DT', 'name': '定远县', 'city': '滁州市', 'lat': 32.53, 'lon': 117.68},
        {'code': 'CZ_FY', 'name': '凤阳县', 'city': '滁州市', 'lat': 32.87, 'lon': 117.56},
        {'code': 'CZ_TT', 'name': '天长市', 'city': '滁州市', 'lat': 32.69, 'lon': 119.00},
        {'code': 'CZ_MS', 'name': '明光市', 'city': '滁州市', 'lat': 32.78, 'lon': 117.99},
        {'code': 'FY_YZ', 'name': '颍州区', 'city': '阜阳市', 'lat': 32.89, 'lon': 115.82},
        {'code': 'FY_YD', 'name': '颍东区', 'city': '阜阳市', 'lat': 32.91, 'lon': 115.86},
        {'code': 'FY_YQ', 'name': '颍泉区', 'city': '阜阳市', 'lat': 32.93, 'lon': 115.81},
        {'code': 'FY_LG', 'name': '临泉县', 'city': '阜阳市', 'lat': 33.06, 'lon': 115.26},
        {'code': 'FY_TY', 'name': '太和县', 'city': '阜阳市', 'lat': 33.16, 'lon': 115.62},
        {'code': 'FY_FS', 'name': '阜南县', 'city': '阜阳市', 'lat': 32.64, 'lon': 115.59},
        {'code': 'FY_YS', 'name': '颍上县', 'city': '阜阳市', 'lat': 32.63, 'lon': 116.26},
        {'code': 'FY_JS', 'name': '界首市', 'city': '阜阳市', 'lat': 33.25, 'lon': 115.37},
        {'code': 'SS_DH', 'name': '埇桥区', 'city': '宿州市', 'lat': 33.63, 'lon': 116.97},
        {'code': 'SS_DN', 'name': '砀山县', 'city': '宿州市', 'lat': 34.43, 'lon': 116.36},
        {'code': 'SS_Xiao', 'name': '萧县', 'city': '宿州市', 'lat': 34.19, 'lon': 116.79},
        {'code': 'SS_LG', 'name': '灵璧县', 'city': '宿州市', 'lat': 33.55, 'lon': 117.44},
        {'code': 'SS_SI', 'name': '泗县', 'city': '宿州市', 'lat': 33.49, 'lon': 117.41},
        {'code': 'LA_JH', 'name': '金安区', 'city': '六安市', 'lat': 31.74, 'lon': 116.51},
        {'code': 'LA_YA', 'name': '裕安区', 'city': '六安市', 'lat': 31.74, 'lon': 116.48},
        {'code': 'LA_HH', 'name': '霍邱县', 'city': '六安市', 'lat': 32.36, 'lon': 116.28},
        {'code': 'LA_SC', 'name': '舒城县', 'city': '六安市', 'lat': 31.46, 'lon': 116.95},
        {'code': 'LA_JS', 'name': '金寨县', 'city': '六安市', 'lat': 31.73, 'lon': 115.93},
        {'code': 'LA_HS', 'name': '霍山县', 'city': '六安市', 'lat': 31.39, 'lon': 116.33},
        {'code': 'CZ_NC', 'name': '宣州区', 'city': '宣城市', 'lat': 30.95, 'lon': 118.76},
        {'code': 'CZ_NJ', 'name': '宁国市', 'city': '宣城市', 'lat': 30.63, 'lon': 118.98},
        {'code': 'CZ_LZ', 'name': '郎溪县', 'city': '宣城市', 'lat': 31.13, 'lon': 119.18},
        {'code': 'CZ_GD', 'name': '广德县', 'city': '宣城市', 'lat': 30.89, 'lon': 119.42},
        {'code': 'CZ_JX', 'name': '泾县', 'city': '宣城市', 'lat': 30.69, 'lon': 118.42},
        {'code': 'CZ_JD', 'name': '绩溪县', 'city': '宣城市', 'lat': 30.07, 'lon': 118.60},
        {'code': 'CZ_JX2', 'name': '旌德县', 'city': '宣城市', 'lat': 30.29, 'lon': 118.54},
        {'code': 'CQ_YC', 'name': '贵池区', 'city': '池州市', 'lat': 30.66, 'lon': 117.52},
        {'code': 'CQ_DZ', 'name': '东至县', 'city': '池州市', 'lat': 30.10, 'lon': 117.03},
        {'code': 'CQ_ST', 'name': '石台县', 'city': '池州市', 'lat': 30.21, 'lon': 117.49},
        {'code': 'CQ_QY', 'name': '青阳县', 'city': '池州市', 'lat': 30.64, 'lon': 117.85},
    ],
    "jiangsu": [
        {'code': 'NJ_XW', 'name': '玄武区', 'city': '南京市', 'lat': 32.05, 'lon': 118.79},
        {'code': 'NJ_QH', 'name': '秦淮区', 'city': '南京市', 'lat': 32.02, 'lon': 118.79},
        {'code': 'NJ_JY', 'name': '建邺区', 'city': '南京市', 'lat': 32.00, 'lon': 118.73},
        {'code': 'NJ_GL', 'name': '鼓楼区', 'city': '南京市', 'lat': 32.07, 'lon': 118.77},
        {'code': 'NJ_PK', 'name': '浦口区', 'city': '南京市', 'lat': 32.07, 'lon': 118.63},
        {'code': 'NJ_QS', 'name': '栖霞区', 'city': '南京市', 'lat': 32.10, 'lon': 118.89},
        {'code': 'NJ_YH', 'name': '雨花台区', 'city': '南京市', 'lat': 31.99, 'lon': 118.78},
        {'code': 'NJ_JN', 'name': '江宁区', 'city': '南京市', 'lat': 31.95, 'lon': 118.84},
        {'code': 'NJ_LH', 'name': '六合区', 'city': '南京市', 'lat': 32.34, 'lon': 118.84},
        {'code': 'NJ_SS', 'name': '溧水区', 'city': '南京市', 'lat': 31.65, 'lon': 119.03},
        {'code': 'NJ_GC', 'name': '高淳区', 'city': '南京市', 'lat': 31.33, 'lon': 118.89},
        {'code': 'SZ_QQ', 'name': '虎丘区', 'city': '苏州市', 'lat': 31.30, 'lon': 120.57},
        {'code': 'SZ_WZ', 'name': '吴中区', 'city': '苏州市', 'lat': 31.26, 'lon': 120.63},
        {'code': 'SZ_XC', 'name': '相城区', 'city': '苏州市', 'lat': 31.37, 'lon': 120.64},
        {'code': 'SZ_GS', 'name': '姑苏区', 'city': '苏州市', 'lat': 31.31, 'lon': 120.62},
        {'code': 'SZ_WJ', 'name': '吴江区', 'city': '苏州市', 'lat': 30.99, 'lon': 120.65},
        {'code': 'SZ_CS', 'name': '常熟市', 'city': '苏州市', 'lat': 31.65, 'lon': 120.75},
        {'code': 'SZ_ZJ', 'name': '张家港市', 'city': '苏州市', 'lat': 31.87, 'lon': 120.56},
        {'code': 'SZ_KS', 'name': '昆山市', 'city': '苏州市', 'lat': 31.39, 'lon': 120.98},
        {'code': 'SZ_TS', 'name': '太仓市', 'city': '苏州市', 'lat': 31.45, 'lon': 121.13},
        {'code': 'WX_LC', 'name': '梁溪区', 'city': '无锡市', 'lat': 31.56, 'lon': 120.30},
        {'code': 'WX_XZ', 'name': '锡山区', 'city': '无锡市', 'lat': 31.59, 'lon': 120.36},
        {'code': 'WX_HS', 'name': '惠山区', 'city': '无锡市', 'lat': 31.68, 'lon': 120.30},
        {'code': 'WX_BH', 'name': '滨湖区', 'city': '无锡市', 'lat': 31.52, 'lon': 120.27},
        {'code': 'WX_XW', 'name': '新吴区', 'city': '无锡市', 'lat': 31.53, 'lon': 120.36},
        {'code': 'WX_JY', 'name': '江阴市', 'city': '无锡市', 'lat': 31.92, 'lon': 120.28},
        {'code': 'WX_YX', 'name': '宜兴市', 'city': '无锡市', 'lat': 31.34, 'lon': 119.82},
        {'code': 'CZ_TS', 'name': '天宁区', 'city': '常州市', 'lat': 31.78, 'lon': 119.97},
        {'code': 'CZ_ZL', 'name': '钟楼区', 'city': '常州市', 'lat': 31.78, 'lon': 119.90},
        {'code': 'CZ_XB', 'name': '新北区', 'city': '常州市', 'lat': 31.83, 'lon': 119.97},
        {'code': 'CZ_WJ', 'name': '武进区', 'city': '常州市', 'lat': 31.70, 'lon': 119.95},
        {'code': 'CZ_JT', 'name': '金坛区', 'city': '常州市', 'lat': 31.74, 'lon': 119.60},
        {'code': 'CZ_LY', 'name': '溧阳市', 'city': '常州市', 'lat': 31.42, 'lon': 119.48},
        {'code': 'NT_CC', 'name': '崇川区', 'city': '南通市', 'lat': 31.98, 'lon': 120.60},
        {'code': 'NT_GC', 'name': '港闸区', 'city': '南通市', 'lat': 32.05, 'lon': 120.81},
        {'code': 'NT_TZ', 'name': '通州区', 'city': '南通市', 'lat': 32.07, 'lon': 121.07},
        {'code': 'NT_HA', 'name': '海安县', 'city': '南通市', 'lat': 32.53, 'lon': 120.47},
        {'code': 'NT_RD', 'name': '如东县', 'city': '南通市', 'lat': 32.33, 'lon': 121.19},
        {'code': 'NT_QD', 'name': '启东市', 'city': '南通市', 'lat': 31.81, 'lon': 121.66},
        {'code': 'NT_RG', 'name': '如皋市', 'city': '南通市', 'lat': 32.26, 'lon': 120.57},
        {'code': 'NT_HM', 'name': '海门市', 'city': '南通市', 'lat': 31.89, 'lon': 121.18},
        {'code': 'YZ_GL', 'name': '广陵区', 'city': '扬州市', 'lat': 32.40, 'lon': 119.42},
        {'code': 'YZ_ZJ', 'name': '邗江区', 'city': '扬州市', 'lat': 32.38, 'lon': 119.40},
        {'code': 'YZ_JD', 'name': '江都区', 'city': '扬州市', 'lat': 32.43, 'lon': 119.57},
        {'code': 'YZ_BY', 'name': '宝应县', 'city': '扬州市', 'lat': 33.24, 'lon': 119.30},
        {'code': 'YZ_YZ', 'name': '仪征市', 'city': '扬州市', 'lat': 32.27, 'lon': 119.18},
        {'code': 'YZ_GY', 'name': '高邮市', 'city': '扬州市', 'lat': 32.46, 'lon': 119.46},
        {'code': 'ZJ_JK', 'name': '京口区', 'city': '镇江市', 'lat': 32.20, 'lon': 119.47},
        {'code': 'ZJ_RC', 'name': '润州区', 'city': '镇江市', 'lat': 32.21, 'lon': 119.43},
        {'code': 'ZJ_DF', 'name': '丹徒区', 'city': '镇江市', 'lat': 32.13, 'lon': 119.43},
        {'code': 'ZJ_DY', 'name': '丹阳市', 'city': '镇江市', 'lat': 32.00, 'lon': 119.60},
        {'code': 'ZJ_YZ', 'name': '扬中市', 'city': '镇江市', 'lat': 32.24, 'lon': 119.83},
        {'code': 'ZJ_JR', 'name': '句容市', 'city': '镇江市', 'lat': 31.95, 'lon': 119.17},
        {'code': 'YC_TF', 'name': '亭湖区', 'city': '盐城市', 'lat': 33.15, 'lon': 120.16},
        {'code': 'YC_YD', 'name': '盐都区', 'city': '盐城市', 'lat': 33.13, 'lon': 120.15},
        {'code': 'YC_DZ', 'name': '大丰区', 'city': '盐城市', 'lat': 33.20, 'lon': 120.45},
        {'code': 'YC_XS', 'name': '响水县', 'city': '盐城市', 'lat': 34.20, 'lon': 119.58},
        {'code': 'YC_BH', 'name': '滨海县', 'city': '盐城市', 'lat': 33.99, 'lon': 119.82},
        {'code': 'YC_FN', 'name': '阜宁县', 'city': '盐城市', 'lat': 33.78, 'lon': 119.80},
        {'code': 'YC_SN', 'name': '射阳县', 'city': '盐城市', 'lat': 33.78, 'lon': 120.26},
        {'code': 'YC_JH', 'name': '建湖县', 'city': '盐城市', 'lat': 33.47, 'lon': 119.80},
        {'code': 'YC_DS', 'name': '东台市', 'city': '盐城市', 'lat': 32.85, 'lon': 120.13},
        {'code': 'YC_YS', 'name': '建湖县', 'city': '盐城市', 'lat': 33.47, 'lon': 119.80},
    ],
    "shanghai": [
        {'code': 'SH_HB', 'name': '黄浦区', 'city': '上海市', 'lat': 31.23, 'lon': 121.49},
        {'code': 'SH_XH', 'name': '徐汇区', 'city': '上海市', 'lat': 31.18, 'lon': 121.44},
        {'code': 'SH_CZ', 'name': '长宁区', 'city': '上海市', 'lat': 31.22, 'lon': 121.42},
        {'code': 'SH_JA', 'name': '静安区', 'city': '上海市', 'lat': 31.23, 'lon': 121.45},
        {'code': 'SH_PT', 'name': '普陀区', 'city': '上海市', 'lat': 31.25, 'lon': 121.40},
        {'code': 'SH_ZB', 'name': '闸北区', 'city': '上海市', 'lat': 31.25, 'lon': 121.46},
        {'code': 'SH_HP', 'name': '虹口区', 'city': '上海市', 'lat': 31.26, 'lon': 121.49},
        {'code': 'SH_YK', 'name': '杨浦区', 'city': '上海市', 'lat': 31.28, 'lon': 121.53},
        {'code': 'SH_PD', 'name': '浦东新区', 'city': '上海市', 'lat': 31.22, 'lon': 121.54},
        {'code': 'SH_MH', 'name': '闵行区', 'city': '上海市', 'lat': 31.11, 'lon': 121.38},
        {'code': 'SH_BS', 'name': '宝山区', 'city': '上海市', 'lat': 31.40, 'lon': 121.49},
        {'code': 'SH_JD', 'name': '嘉定区', 'city': '上海市', 'lat': 31.38, 'lon': 121.27},
        {'code': 'SH_JS', 'name': '金山区', 'city': '上海市', 'lat': 30.84, 'lon': 121.34},
        {'code': 'SH_SJ', 'name': '松江区', 'city': '上海市', 'lat': 31.01, 'lon': 121.23},
        {'code': 'SH_QP', 'name': '青浦区', 'city': '上海市', 'lat': 31.15, 'lon': 121.12},
        {'code': 'SH_FX', 'name': '奉贤区', 'city': '上海市', 'lat': 30.92, 'lon': 121.47},
        {'code': 'SH_CM', 'name': '崇明区', 'city': '上海市', 'lat': 31.62, 'lon': 121.40},
    ],
}


def get_conn():
    return psycopg2.connect(**PG)


def safe_float(v):
    if v is None:
        return None
    try:
        return float(v)
    except (ValueError, TypeError):
        return None


def fetch_open_meteo(lat, lon, days=3):
    """调用 Open-Meteo API 获取逐小时+逐日数据"""
    end_date = date.today()
    start_date = end_date - timedelta(days=days)

    url = (
        f"https://api.open-meteo.com/v1/forecast"
        f"?latitude={lat}&longitude={lon}"
        f"&hourly=temperature_2m,apparent_temperature,relative_humidity_2m,dew_point_2m,"
        f"precipitation,rain,snowfall,snow_depth,wind_speed_10m,wind_speed_100m,"
        f"wind_direction_10m,wind_direction_100m,wind_gusts_10m,pressure_msl,"
        f"surface_pressure,cloud_cover,cloud_cover_low,cloud_cover_mid,cloud_cover_high,"
        f"shortwave_radiation,direct_radiation,diffuse_radiation,uv_index,visibility,"
        f"weather_code,is_day,et0_fao_evapotranspiration,soil_temperature_0_to_7cm,"
        f"soil_moisture_0_to_7cm"
        f"&daily=temperature_2m_max,temperature_2m_min,temperature_2m_mean,"
        f"relative_humidity_2m_mean,precipitation_sum,rain_sum,snowfall_sum,"
        f"et0_fao_evapotranspiration,wind_speed_10m_max"
        f"&start_date={start_date.isoformat()}&end_date={end_date.isoformat()}"
        f"&timezone=Asia%2FShanghai"
    )
    req = urllib.request.Request(url)
    with urllib.request.urlopen(req, timeout=30) as resp:
        return json.loads(resp.read().decode())


def upsert_hourly(loc, data):
    """写入 md_weather_hourly"""
    hourly = data.get("hourly", {})
    times = hourly.get("time", [])
    if not times:
        return 0

    # 先删除该地点旧的小时数据（避免重复）
    conn = get_conn()
    cur = conn.cursor()
    adcode = loc["code"]
    cur.execute(
        "DELETE FROM md_weather_hourly WHERE adcode = %s AND obs_time >= NOW() - INTERVAL '7 days'",
        (adcode,),
    )
    conn.commit()
    cur.close()
    conn.close()

    rows = []
    for i, t in enumerate(times):
        rows.append({
            "adcode": adcode,
            "district_name": loc["name"],
            "city_name": loc.get("city", ""),
            "province_name": loc.get("province", ""),
            "lat": loc["lat"],
            "lon": loc["lon"],
            "obs_time": t,
            "temperature_2m": safe_float(hourly.get("temperature_2m", [None])[i]) if i < len(hourly.get("temperature_2m", [])) else None,
            "apparent_temperature": safe_float(hourly.get("apparent_temperature", [None])[i]) if i < len(hourly.get("apparent_temperature", [])) else None,
            "relative_humidity_2m": safe_float(hourly.get("relative_humidity_2m", [None])[i]) if i < len(hourly.get("relative_humidity_2m", [])) else None,
            "wind_speed_10m": safe_float(hourly.get("wind_speed_10m", [None])[i]) if i < len(hourly.get("wind_speed_10m", [])) else None,
            "wind_speed_100m": safe_float(hourly.get("wind_speed_100m", [None])[i]) if i < len(hourly.get("wind_speed_100m", [])) else None,
            "wind_direction_10m": int(hourly["wind_direction_10m"][i]) if i < len(hourly.get("wind_direction_10m", [])) and hourly["wind_direction_10m"][i] is not None else None,
            "wind_direction_100m": int(hourly["wind_direction_100m"][i]) if i < len(hourly.get("wind_direction_100m", [])) and hourly["wind_direction_100m"][i] is not None else None,
            "pressure_msl": safe_float(hourly.get("pressure_msl", [None])[i]) if i < len(hourly.get("pressure_msl", [])) else None,
            "precipitation": safe_float(hourly.get("precipitation", [None])[i]) if i < len(hourly.get("precipitation", [])) else None,
        })

    if not rows:
        return 0

    conn = get_conn()
    cur = conn.cursor()
    cols = list(rows[0].keys())
    col_str = ", ".join(cols)
    sql = f"INSERT INTO md_weather_hourly ({col_str}) VALUES %s"
    vals = [tuple(r[c] for c in cols) for r in rows]
    try:
        execute_values(cur, sql, vals)
        conn.commit()
        return len(rows)
    except Exception as e:
        conn.rollback()
        print(f"    [!] hourly upsert 失败: {e}")
        return 0
    finally:
        cur.close()
        conn.close()


def upsert_hydrology(loc, data):
    """写入 md_weather_hydrology_daily"""
    daily = data.get("daily", {})
    dates = daily.get("time", [])
    if not dates:
        return 0

    rows = []
    for i, d in enumerate(dates):
        rows.append({
            "location_code": loc["code"],
            "location_name": loc["name"],
            "lat": loc["lat"],
            "lon": loc["lon"],
            "obs_date": d,
            "temp_mean": safe_float(daily.get("temperature_2m_mean", [None])[i]) if i < len(daily.get("temperature_2m_mean", [])) else None,
            "humidity_mean": safe_float(daily.get("relative_humidity_2m_mean", [None])[i]) if i < len(daily.get("relative_humidity_2m_mean", [])) else None,
            "precipitation_sum": safe_float(daily.get("precipitation_sum", [None])[i]) if i < len(daily.get("precipitation_sum", [])) else None,
            "rain_sum": safe_float(daily.get("rain_sum", [None])[i]) if i < len(daily.get("rain_sum", [])) else None,
            "et0_evapotranspiration": safe_float(daily.get("et0_fao_evapotranspiration", [None])[i]) if i < len(daily.get("et0_fao_evapotranspiration", [])) else None,
            "wind_speed_10m_mean": safe_float(daily.get("wind_speed_10m_max", [None])[i]) if i < len(daily.get("wind_speed_10m_max", [])) else None,
        })

    conn = get_conn()
    cur = conn.cursor()
    # delete existing for this location + date range
    cur.execute(
        "DELETE FROM md_weather_hydrology_daily WHERE location_code = %s AND obs_date >= %s",
        (loc["code"], dates[0]),
    )
    conn.commit()

    cols = list(rows[0].keys())
    col_str = ", ".join(cols)
    sql = f"INSERT INTO md_weather_hydrology_daily ({col_str}) VALUES %s"
    vals = [tuple(r[c] for c in cols) for r in rows]
    try:
        execute_values(cur, sql, vals)
        conn.commit()
        return len(rows)
    except Exception as e:
        conn.rollback()
        print(f"    [!] hydrology upsert 失败: {e}")
        return 0
    finally:
        cur.close()
        conn.close()


def fetch_province(prov_name, locations):
    total_h = 0
    total_d = 0
    for i, loc in enumerate(locations):
        prov_full = {"fujian": "福建省", "jiangxi": "江西省", "anhui": "安徽省",
                     "jiangsu": "江苏省", "shanghai": "上海市"}.get(prov_name, prov_name)
        loc["province"] = prov_full
        try:
            data = fetch_open_meteo(loc["lat"], loc["lon"])
            h = upsert_hourly(loc, data)
            d = upsert_hydrology(loc, data)
            total_h += h
            total_d += d
            if (i + 1) % 10 == 0:
                print(f"  [{prov_name}] {i+1}/{len(locations)} 已完成")
        except Exception as e:
            print(f"  [!] {loc['name']} 失败: {e}")
        time.sleep(0.1)  # 限速
    print(f"  [{prov_name}] 完成: {total_h} hourly + {total_d} daily rows")
    return total_h, total_d


def main():
    prov_arg = sys.argv[1].lower() if len(sys.argv) > 1 else "all"
    if prov_arg == "all":
        for prov, locs in PROVINCES.items():
            print(f"\n=== {prov} ({len(locs)} 区县) ===")
            fetch_province(prov, locs)
    elif prov_arg in PROVINCES:
        print(f"\n=== {prov_arg} ({len(PROVINCES[prov_arg])} 区县) ===")
        fetch_province(prov_arg, PROVINCES[prov_arg])
    else:
        print(f"未知省份: {prov_arg}，可选: {', '.join(PROVINCES.keys())}")
        sys.exit(1)
    print("\n[WEATHER] 采集完成")


if __name__ == "__main__":
    main()
