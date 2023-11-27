import subprocess
import time

go_program_path = '/path/to/your/go/program'
num_runs = 100
cpu_times = []

for i in range(num_runs):
    start_time = time.process_time()
    subprocess.run(['go', 'run', go_program_path], stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    end_time = time.process_time()
    cpu_time = end_time - start_time
    cpu_times.append(cpu_time)

average_cpu_time = sum(cpu_times) / num_runs
output_file_path = 'average_cpu_time.txt'

with open(output_file_path, 'w') as output_file:
    output_file.write(f'Average CPU Time: {average_cpu_time} seconds\n')

print(f'Average CPU Time: {average_cpu_time} seconds')
print(f'Results saved to {output_file_path}')
