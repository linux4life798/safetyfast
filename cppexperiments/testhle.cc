/**
 * @brief This program tests Intel HLE primitives against other mutex primitives
 *        in a high sparsity environment, where transactions should excel.
 *
 * The idea is that transactional memory should excel when multiple threads are
 * accessing different memory locations, that do not conflict.
 *
 *
 * @file testhle.cc
 * @author Craig Hesling
 * @date 2018-04-01
 */
#include <stdio.h>
#include <stdlib.h>
#include <iostream>
#include <string>
#include <thread>
#include <chrono>
#include <vector>
#include <set>
#include <mutex>
#include <immintrin.h>

using namespace std;
using namespace std::chrono;

#define PERCENT(numerator,denominator) \
		(((long double)(long long)(numerator)) / ((long double)(long long)(denominator))) * ((long double)100)

// #define DEFAULT_LOOP_ATTEMPTS 10000
#define DEFAULT_LOOP_ATTEMPTS 4000000
#define DEFAULT_NUM_THREADS   6
#define DEFAULT_ROUNDS        10

// Referencing any of the children types using base class is slower than the direct type
class CustomMutex {
public:
	virtual inline void lock() {}
	virtual inline void unlock() {}
	virtual string gettype() = 0;
	virtual ~CustomMutex() {}
};


/**
 * @brief The anti-mutex that doesn't have any effect.
 * The methods \a lock and \a unlock simply return immediately.
 */
class NoMutex : public CustomMutex {
public:
	string gettype() {
		return "NoMutex";
	}
};

/**
 * @brief Simply a proxy for the standard \a std::mutex.
 */
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

/**
 * @brief A very basic spin mutex implementation.
 */
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

template <class CustomMutexType>
class MySet {
	size_t numbins;
	int *values;
	CustomMutexType m;
public:

	// allocate numbins bins which will be randomly touched by threads
	MySet(size_t numbins) :numbins(numbins){
		// Enforce CustomMutexType is a descendant of CustomMutex
		static_assert(std::is_base_of<CustomMutex, CustomMutexType>::value, "Template type is not derived from CustomMutex");
		values = new int[numbins];
	}
	~MySet() {
		delete[] values;
	}

	// increment the value in bin binindex
	void touch(size_t binindex) {
		size_t bin = binindex % numbins;
		m.lock();
		values[bin]++;
		m.unlock();
	}

	size_t len() {
		return numbins;
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


template <class CustomMutexType>
void thread_func(params *par, result *res, MySet<CustomMutexType> *s) {
	unsigned long long loops = par->loops;
	auto vals = &par->vals;

	for (unsigned long long i = 0; i < loops; i++) {
		s->touch((*vals)[i]);
	}
}

template <class CustomMutexType>
long double runtest(int num_threads, int loops, int rounds) {
	printf("# Spawning %d threads to do %d loops\n", num_threads, loops);

	thread threads[num_threads];
	params    *par[num_threads];
	result    *res[num_threads];

	for (int tid = 0; tid < num_threads; tid++) {
		par[tid] = new params(tid, loops);
		res[tid] = new result;
		res[tid]->statuses = new unsigned[par[tid]->loops];
	}

	long double sum = 0.0;
	for (int r = 0; r < rounds; r++) {
		MySet<CustomMutexType> s(100000);
		auto start = high_resolution_clock::now();
		for (int tid = 0; tid < num_threads; tid++) {
			threads[tid] = thread(thread_func<CustomMutexType>, par[tid], res[tid], &s);
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

	return avg_dur;
}

int main(int argc, char *argv[]) {
	int num_threads = DEFAULT_NUM_THREADS;
	int loops       = DEFAULT_LOOP_ATTEMPTS;
	int rounds      = DEFAULT_ROUNDS;

	// Search for help flags
	for (int i = 1; i < argc; i++) {
		string arg(argv[i]);
		if (arg.compare("-h")==0 || arg.compare("--help")==0) {
			cout << "Usage: testhle [num_threads] [loops] [rounds]" << endl;
			exit(0);
		}
	}

	if (argc > 1) {
		num_threads = atoi(argv[1]);
	}

	if (argc > 2) {
		loops = atoi(argv[2]);
	}

	if (argc > 3) {
		rounds = atoi(argv[3]);
	}

	runtest<NoMutex>(num_threads, loops, rounds);
	printf("\n");
	runtest<SystemMutex>(num_threads, loops, rounds);
	printf("\n");
	runtest<SpinMutex>(num_threads, loops, rounds);
	printf("\n");
	runtest<SpinHLEMutex>(num_threads, loops, rounds);

	return 0;
}
