/**
 * @brief This program tests the collision rate when using n threads and memory transactions
 *
 * @file testhle2.cc
 * @author Craig Hesling
 * @date 2018-04-01
 */
#include <stdio.h>
#include <stdlib.h>
#include <iostream>
#include <thread>
#include <chrono>
#include <vector>
// #include <map>
#include <set>
#include <mutex>
#include <immintrin.h>

using namespace std;
using namespace std::chrono;

#define PERCENT(numerator,denominator) \
		(((long double)(long long)(numerator)) / ((long double)(long long)(denominator))) * ((long double)100)

// #define DEFAULT_LOOP_ATTEMPTS 10000
#define DEFAULT_LOOP_ATTEMPTS 4000000
#define DEFAULT_NUM_THREADS   8
#define DEFAULT_ROUNDS        10

// Referencing any of the children types using base class is slower than the direct type
class CustomMutex {
public:
	virtual inline void lock() {}
	virtual inline void unlock() {}
	virtual string gettype() = 0;
	virtual ~CustomMutex() {}
};

class NoMutex : public CustomMutex {
	string gettype() {
		return "NoMutex";
	}
};

class SystemMutex : public CustomMutex {
	std::mutex m;
public:
	inline
	void lock() {
		m.lock();
	}
	inline
	void unlock() {
		m.unlock();
	}
	string gettype() {
		return "SystemMutex";
	}
};

class SpinMutex : public CustomMutex {
	int val = 0;
public:
	inline
	void lock() {
		do {
			while(val != 0) {
				// asm volatile("pause\n": : :"memory");
				_mm_pause();
			}
		} while(__sync_lock_test_and_set(&val, 1) != 0);
	}

	inline
	void unlock() {
		__sync_lock_release(&val);
	}

	string gettype() {
		return "SpinMutex";
	}
};

class SpinHLEMutex : public CustomMutex {
	int val = 0;
public:
	inline
	void lock() {
		do {
			while(val != 0) {
				_mm_pause();
			}
		} while(__atomic_exchange_n(&val, 1, __ATOMIC_ACQUIRE|__ATOMIC_HLE_ACQUIRE));
	}

	inline
	void unlock() {
		/* Free lock with lock elision */
		__atomic_store_n(&val, 0, __ATOMIC_RELEASE|__ATOMIC_HLE_RELEASE);
	}

	string gettype() {
		return "SpinHLEMutex";
	}
};

class MySet {
	size_t bins;
	int *values;
	// mutex m;
	// SystemMutex m;
	// SpinMutex m;
	SpinHLEMutex m;
	// NoMutex m;
	// MutexType m;
	// CustomMutex *m;
	// bool fallback;
public:
	MySet(size_t bins) :bins(bins){
		// locks = new SpinHLEMutex[bins];
		values = new int[bins];
	}
	~MySet() {
		// delete[] locks;
		delete[] values;
	}
	void put(unsigned long long val) {
		unsigned long long bin = val % bins;
		m.lock();
		values[bin]++;
		m.unlock();
	}
	size_t len() {
		return bins;
	}
	string getlocktype() {
		// return "RTM with SpinHLEMutex fallback";
		return m.gettype();
	}
};


struct params {
	int bins;
	int tid;
	size_t loops;
	int *vals;

	params(int tid, size_t loops) : tid(tid), loops(loops) {
		vals = new int[loops];
		for (size_t i = 0; i < loops; i++) {
			vals[i] = rand();
		}
	}
	~params() {
		delete[] vals;
	}
};

struct result {
	unsigned long long success;
	unsigned long long failure;
	unsigned          *statuses;
};

unsigned long long int counter = 0;

#define print_status_flag(tid,loopi,status,flag) \
			if (status & flag)                   \
				printf("[Thread %2d] loop %d - %s\n", tid, loopi, #flag)

void thread_func(params *par, result *res, MySet *s) {
	unsigned long long loops = par->loops;
	auto vals = &par->vals;

	for (unsigned long long i = 0; i < loops; i++) {
		s->put((*vals)[i]);
	}
}

int main(int argc, char *argv[]) {
	int num_threads = DEFAULT_NUM_THREADS;
	int loops = DEFAULT_LOOP_ATTEMPTS;
	int rounds = DEFAULT_ROUNDS;

	if (argc > 1) {
		num_threads = atoi(argv[1]);
	}

	if (argc > 2) {
		loops = atoi(argv[2]);
	}

	if (argc > 3) {
		rounds = atoi(argv[3]);
	}

	printf("# Spawning %d threads to do %d loops\n", num_threads, loops);

	thread threads[num_threads];
	params    *par[num_threads];
	result    *res[num_threads];

	for (int tid = 0; tid < num_threads; tid++) {
		par[tid] = new params(tid, loops);
		res[tid] = new result;
		res[tid]->statuses = new unsigned[par[tid]->loops];
	}

	// const unsigned long long total = num_threads*loops;
	// unsigned long long success = 0;
	// unsigned long long failure = 0;

	long double sum = 0.0;
	for (int r = 0; r < rounds; r++) {
		// SpinHLEMutex m;
		// CustomMutex *mp = &m;
		MySet s(100000);
		auto start = high_resolution_clock::now();
		for (int tid = 0; tid < num_threads; tid++) {
			threads[tid] = thread(thread_func, par[tid], res[tid], &s);
		}
		for (int tid = 0; tid < num_threads; tid++) {
			threads[tid].join();
		}
		auto stop = high_resolution_clock::now();

		auto dur = duration_cast<duration<long double>>(stop - start).count();
		sum += dur;

		printf("%.4Lf ms - %ld elements - using %s\n", dur*1000, s.len(), s.getlocktype().c_str());
	}
	auto avg_dur = (sum / ((long double)rounds));
	printf("Average %.4Lf ms over %d rounds\n", avg_dur * 1000.0, rounds);
	printf("Average %.4Lf us per item\n", (avg_dur / (long double)(num_threads*loops)) * 1000.0 * 1000.0);


	// for (int tid = 0; tid < num_threads; tid++) {
	// 	success += res[tid]->success;
	// 	failure += res[tid]->failure;
	// 	printf("TH%d - Success=%Lu | Failed=%Lu | Total=%Lu | SuccessRate=%Lf%% | FailureRate=%Lf%%\n",
	// 		tid, res[tid]->success, res[tid]->failure, res[tid]->success+res[tid]->failure,
	// 		PERCENT(res[tid]->success, loops), PERCENT(res[tid]->failure, loops));
	// }

	// printf("MASTER - %Ld\n", counter);

	// printf("Total Attempts %Lu | Total Success %Lu | Total Failed %Lu\n", (unsigned long long)(total), success, failure);
	// printf("Success Rate %Lu / %Lu = %Lf%%\n", success, (unsigned long long)(total), PERCENT(success, total));
	// printf("Failure Rate %Lu / %Lu = %Lf%%\n", failure, (unsigned long long)(total), PERCENT(failure, total) );
	// cout << "Set has " << s.len() << " elements" << endl;
	return 0;
}
