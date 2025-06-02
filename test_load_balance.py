#!/usr/bin/env python3
import requests
import concurrent.futures
import time
from collections import Counter
import statistics
import json

# 配置
URL = "http://localhost/api/v1/communityPost/posts/guest"
# URL = "http://192.168.49.2:30081/api/v1/communityPost/posts/guest"
TOTAL_REQUESTS = 10000
MAX_WORKERS = 100

def make_request(request_id):
    """发送单个请求并返回结果"""
    start_time = time.time()
    try:
        response = requests.get(URL, timeout=10)
        response_time = time.time() - start_time

        # 尝试从响应头获取实例ID
        instance_id = response.headers.get('X-Instance-ID', 'unknown')

        # 如果响应头没有，尝试从响应体获取
        if instance_id == 'unknown':
            try:
                data = response.json()
                instance_id = data.get('instance', 'unknown')
            except:
                pass

        return {
            'request_id': request_id,
            'instance_id': instance_id,
            'status_code': response.status_code,
            'response_time': response_time,
            'success': True
        }
    except Exception as e:
        return {
            'request_id': request_id,
            'instance_id': 'error',
            'status_code': 0,
            'response_time': time.time() - start_time,
            'success': False,
            'error': str(e)
        }

def main():
    print("=" * 60)
    print(f"负载均衡测试")
    print(f"URL: {URL}")
    print(f"总请求数: {TOTAL_REQUESTS}")
    print(f"并发数: {MAX_WORKERS}")
    print("=" * 60)

    # 执行并发请求
    print("\n开始测试...")
    start_time = time.time()

    results = []
    with concurrent.futures.ThreadPoolExecutor(max_workers=MAX_WORKERS) as executor:
        futures = [executor.submit(make_request, i) for i in range(TOTAL_REQUESTS)]

        for future in concurrent.futures.as_completed(futures):
            result = future.result()
            results.append(result)

            # 实时显示进度
            if len(results) % 10 == 0:
                print(f"已完成: {len(results)}/{TOTAL_REQUESTS}")

    total_time = time.time() - start_time

    # 分析结果
    print("\n" + "=" * 60)
    print("测试结果:")
    print("=" * 60)

    # 请求分布
    instance_counter = Counter(r['instance_id'] for r in results)
    print("\n请求分布:")
    for instance, count in sorted(instance_counter.items()):
        percentage = (count / TOTAL_REQUESTS) * 100
        print(f"  实例 {instance}: {count} 次 ({percentage:.1f}%)")

    # 成功率
    successful_requests = sum(1 for r in results if r['success'])
    success_rate = (successful_requests / TOTAL_REQUESTS) * 100
    print(f"\n成功率: {success_rate:.1f}% ({successful_requests}/{TOTAL_REQUESTS})")

    # 响应时间统计
    response_times = [r['response_time'] for r in results if r['success']]
    if response_times:
        print(f"\n响应时间统计:")
        print(f"  平均: {statistics.mean(response_times):.3f}s")
        print(f"  中位数: {statistics.median(response_times):.3f}s")
        print(f"  最小: {min(response_times):.3f}s")
        print(f"  最大: {max(response_times):.3f}s")
        if len(response_times) > 1:
            print(f"  标准差: {statistics.stdev(response_times):.3f}s")

    # 状态码分布
    status_counter = Counter(r['status_code'] for r in results)
    print(f"\n状态码分布:")
    for status, count in sorted(status_counter.items()):
        print(f"  {status}: {count} 次")

    # 错误信息
    errors = [r for r in results if not r['success']]
    if errors:
        print(f"\n错误信息 ({len(errors)} 个):")
        for err in errors[:5]:  # 只显示前5个错误
            print(f"  - {err.get('error', 'Unknown error')}")

    print(f"\n总测试时间: {total_time:.2f}s")
    print(f"平均 QPS: {TOTAL_REQUESTS / total_time:.2f}")
    print("=" * 60)

if __name__ == "__main__":
    main()